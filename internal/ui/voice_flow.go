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
// Updates a single in-place line cleanly on stdout without duplicate messages or screen clutter.
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

		// State: RECORDING
		recIndicator := NewVoiceIndicator("Recording audio (Speak now - Press [ENTER] to stop)")
		recIndicator.Start()

		recSession, err := activeEng.StartRecording(ctx, nil)

		var transcribed string
		if err != nil {
			recIndicator.StopFailure("Failed to start recording")
			fmt.Printf("%sMicrophone unavailable: %v%s\n", th.Error, err, th.Reset)
			fmt.Println("Ensure Termux:API has microphone permissions granted in Android Settings.")
		} else {
			// Listen asynchronously on stdin for user pressing [ENTER] or typing 'cancel'
			stopChan := make(chan string, 1)
			go func() {
				line, _ := reader.ReadString('\n')
				stopChan <- strings.TrimSpace(line)
			}()

			select {
			case userInput := <-stopChan:
				// State: STOPPING - transition in-place without producing new line
				recIndicator.Update("Stopping recording...")
				_ = recSession.Stop()

				if strings.ToLower(userInput) == "cancel" {
					_ = os.Remove(recSession.AudioPath)
					recIndicator.StopFailure("Recording cancelled")
					return VoiceSessionResult{Cancelled: true}
				}
			case <-ctx.Done():
				recIndicator.Update("Recording interrupted...")
				_ = recSession.Stop()
				_ = os.Remove(recSession.AudioPath)
				recIndicator.StopFailure("Recording cancelled by system")
				return VoiceSessionResult{Cancelled: true}
			}

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
			continue

		case 3:
			// Cancel
			fmt.Printf("%s[Voice instruction cancelled]%s\n\n", th.Muted, th.Reset)
			return VoiceSessionResult{Cancelled: true}
		}
	}
}
