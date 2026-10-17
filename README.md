# Tertic

Run AI agents on your Android phone - no laptop needed!

## How to use

1. **Download Termux**: https://github.com/termux/termux-app/releases (get the `apt-android-7` APK for Android 7+)

2. **In Termux, run**:
   ```bash
   pkg update && pkg upgrade
   pkg install git
   git clone https://github.com/Ashed2127/tertic.git
   cd tertic
   bash setup.sh
   ```

3. **Run an agent**:
   ```bash
   source ~/tertic-env/bin/activate
   python ~/tertic/examples/simple_agent.py
   ```

That's it! Check `docs/` for more details.
