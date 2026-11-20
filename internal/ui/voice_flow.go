package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Ashed2127/tertic/internal/theme"
	"github.com/Ashed2127/tertic/internal/voice"
)

// VoiceSessionResult represents the outcome of a voice recording turn
type VoiceSessionResult struct {
	Text      string
	Cancelled bool
	Submitted bool
}

// StartVoiceInputFlow implements the robust voice recording state machine:
// IDLE → RECORDING → STOPPING → RECORDED → TRANSCRIBING → COMPLETED
// Listens for ENTER (to stop), ESC (to cancel), or Ctrl+C (to cancel).
// Guarantees immediate stop and terminal restoration on both Linux and Android Termux.
func StartVoiceInputFlow(ctx context.Context, reader *bufio.Reader, voiceMgr *voice.Manager) VoiceSessionResult {
	th := theme.Current()

	fmt.Printf("\n%s%s╔════════════════════════════════════════════════════════════════╗%s\n", th.Border, th.Bold, th.Reset)
	fmt.Printf("%s║                   TERTIC VOICE MODE                            ║%s\n", th.RecordColor, th.Reset)
	fmt.Printf("%s╚════════════════════════════════════════════════════════════════╝%s\n\n", th.Border, th.Reset)

	activeEng := voiceMgr.ActiveEngine()
	if activeEng == nil {
		fmt.Printf("%s%s[Notice] No active speech-recognition engine detected.%s\n", th.Warning, th.Bold, th.Reset)
		fmt.Println("To enable speech on Termux Android:")
		fmt.Println("  1. pkg install termux-api")
		fmt.Println("  2. Install Termux:API app from F-Droid")
		fmt.Println("  3. Grant Microphone permission in Android settings")
		fmt.Println()
		fmt.Printf("%sFalling back to manual instruction prompt below:%s\n", th.Muted, th.Reset)
		fmt.Printf("%sType instruction (or Enter to cancel):%s ", th.Primary, th.Reset)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			return VoiceSessionResult{Cancelled: true}
		}
		return VoiceSessionResult{Text: text, Submitted: true}
	}

	for {
		fmt.Printf("%sEngine:%s %s%s%s\n", th.Bold, th.Reset, th.Primary, activeEng.Name(), th.Reset)
		fmt.Printf("%sOptions:%s Press [ENTER] to start recording (or type 'cancel' / 'exit'): ", th.Muted, th.Reset)
		initInput, _ := reader.ReadString('\n')
		initTrimmed := strings.ToLower(strings.TrimSpace(initInput))
		if initTrimmed == "cancel" || initTrimmed == "exit" || initTrimmed == "q" {
			fmt.Printf("%s[Voice session cancelled]%s\n\n", th.Muted, th.Reset)
			return VoiceSessionResult{Cancelled: true}
		}

		// If the engine supports live direct capture (e.g. Android Native SpeechRecognizer via Termux:API)
		if activeEng.HasLiveCapture() {
			recIndicator := NewVoiceIndicator("Listening for speech (Speak now - Android SpeechRecognizer)...")
			recIndicator.Start()

			transcribed, err := activeEng.LiveCapture(ctx, func(s string) {
				recIndicator.Update(s)
			})

			if err != nil {
				recIndicator.StopFailure("Voice capture failed")
				fmt.Printf("%sVoice capture error: %v%s\n", th.Error, err, th.Reset)
				fmt.Printf("\n%sTry recording again? (y/n):%s ", th.Warning, th.Reset)
				retryAns, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(retryAns)) == "y" {
					continue
				}
				return VoiceSessionResult{Cancelled: true}
			}

			recIndicator.StopSuccess("Speech recognized successfully.")
			transcribed = strings.TrimSpace(transcribed)
			if transcribed == "" {
				fmt.Printf("%s[No speech recognized]%s\n", th.Warning, th.Reset)
				fmt.Printf("Try again? (y/n): ")
				retryAns, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(retryAns)) == "y" {
					continue
				}
				return VoiceSessionResult{Cancelled: true}
			}

			return reviewAndSubmit(reader, th, transcribed)
		}

		// State: RECORDING (Audio file recording workflow)
		recIndicator := NewVoiceIndicator("Recording audio (Speak now - Press [ENTER] to stop | ESC to cancel)")
		recIndicator.Start()

		recSession, err := activeEng.StartRecording(ctx, nil)

		var transcribed string
		if err != nil {
			recIndicator.StopFailure("Failed to start recording")
			fmt.Printf("%sMicrophone unavailable: %v%s\n", th.Error, err, th.Reset)
			fmt.Println("Ensure Termux:API has microphone permissions granted in Android Settings.")
		} else {
			// Channel for stop/cancel key events
			keyChan := make(chan RawKey, 1)
			doneKeyListen := make(chan struct{})

			// Listen for keystrokes in non-blocking raw mode
			go ListenKeyForRecording(keyChan, doneKeyListen)

			var userKey RawKey
			select {
			case key := <-keyChan:
				userKey = key
			case <-ctx.Done():
				userKey = KeyEscape
			}

			// Signal keyboard listener to exit immediately and restore terminal
			close(doneKeyListen)

			// State: STOPPING - transition in-place without producing new line
			if userKey == KeyEscape || userKey == KeyCtrlC {
				recIndicator.Update("Cancelling recording...")
				_ = recSession.Stop()
				_ = os.Remove(recSession.AudioPath)
				recIndicator.StopFailure("Recording cancelled")
				return VoiceSessionResult{Cancelled: true}
			}

			recIndicator.Update("Stopping recording...")
			_ = recSession.Stop()

			// State: RECORDED (Validate file exists & has size)
			if valErr := voice.ValidateAudioFile(recSession.AudioPath); valErr != nil {
				recIndicator.StopFailure("Recording failed: " + valErr.Error())
				err = valErr
			} else {
				// State: TRANSCRIBING - transition in-place
				recIndicator.Update("Transcribing audio...")
				transcribed, err = activeEng.Transcribe(ctx, recSession.AudioPath, func(s string) {
					recIndicator.Update(s)
				})

				if err == nil {
					// State: COMPLETED
					recIndicator.StopSuccess("Transcription completed.")
					// Clean up temporary recording on success
					_ = os.Remove(recSession.AudioPath)
				} else {
					// State: FAILED (preserve recording file for debugging)
					recIndicator.StopFailure("Transcription failed")
					fmt.Printf("%sTranscription error: %v%s\n", th.Error, err, th.Reset)
					fmt.Printf("%sYour recording has been preserved at: %s%s\n", th.Muted, recSession.AudioPath, th.Reset)
				}
			}
		}

		if err != nil {
			fmt.Printf("\n%sTry recording again? (y/n):%s ", th.Warning, th.Reset)
			retryAns, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(retryAns)) == "y" {
				continue
			}
			return VoiceSessionResult{Cancelled: true}
		}

		transcribed = strings.TrimSpace(transcribed)
		if transcribed == "" {
			fmt.Printf("%s[No speech detected in audio]%s\n", th.Warning, th.Reset)
			fmt.Printf("Record again? (y/n): ")
			retryAns, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(retryAns)) == "y" {
				continue
			}
			return VoiceSessionResult{Cancelled: true}
		}

		res := reviewAndSubmit(reader, th, transcribed)
		if res.Cancelled && res.Text == "__RECORD_AGAIN__" {
			continue
		}
		return res
	}
}

func reviewAndSubmit(reader *bufio.Reader, th theme.Theme, transcribed string) VoiceSessionResult {
	// Transcription Review Box
	fmt.Printf("\n%s──────────────── TRANSCRIPTION ────────────────%s\n", th.Border, th.Reset)
	fmt.Printf(" %s%s\"%s\"%s\n", th.Bold, th.Text, transcribed, th.Reset)
	fmt.Printf("%s────────────────────────────────────────────────%s\n\n", th.Border, th.Reset)

	reviewOpts := []string{
		"1) Submit instruction to Tertic agent (Enter)",
		"2) Edit transcription using keyboard",
		"3) Record again",
		"4) Cancel and return",
	}

	choice := SelectInteractiveWithTheme("Choose Action for Transcribed Instruction:", reviewOpts)

	switch choice {
	case 0:
		// Submit directly
		fmt.Printf("%s✔ Instruction approved & submitted to Tertic.%s\n\n", th.Success, th.Reset)
		return VoiceSessionResult{Text: transcribed, Submitted: true}

	case 1:
		// Edit transcription
		fmt.Printf("\n%sEdit your instruction below (press Enter to confirm):%s\n", th.Primary, th.Reset)
		fmt.Printf("%sOriginal:%s %s\n", th.Muted, th.Reset, transcribed)
		fmt.Printf("%sEdited >%s ", th.Primary, th.Reset)
		edited, _ := reader.ReadString('\n')
		edited = strings.TrimSpace(edited)
		if edited == "" {
			edited = transcribed
		}
		fmt.Printf("%s✔ Instruction updated & submitted.%s\n\n", th.Success, th.Reset)
		return VoiceSessionResult{Text: edited, Submitted: true}

	case 2:
		// Record again
		return VoiceSessionResult{Cancelled: true, Text: "__RECORD_AGAIN__"}

	default:
		// Cancel
		fmt.Printf("%s[Voice instruction cancelled]%s\n\n", th.Muted, th.Reset)
		return VoiceSessionResult{Cancelled: true}
	}
}
