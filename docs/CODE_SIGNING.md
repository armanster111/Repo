# Code Signing

Windows Defender and SmartScreen may warn about unsigned executables downloaded from the internet.

For most users:

1. Download `Muse-Windows.zip`
2. Extract `Muse.exe`
3. If SmartScreen appears, choose **More info** → **Run anyway**

The app is a single self-contained Go binary with no installer dependencies.

## Publishing trusted builds

To remove SmartScreen warnings for your users:

1. Buy a code signing certificate from a trusted CA.
2. Build with `build-windows.bat` or `scripts/package-windows.sh`.
3. Sign with `sign-windows.bat` after setting:
   - `SIGN_CERT` = path to your `.pfx`
   - `SIGN_PASSWORD` = certificate password
4. Upload the signed `dist/Muse.exe` or ZIP to releases.

Unsigned builds are still safe when you build them yourself from source.
