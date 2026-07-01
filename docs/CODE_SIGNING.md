# Code Signing

Windows Defender and SmartScreen often block unsigned executables downloaded from the internet.

To publish a trusted build:

1. Buy a code signing certificate from a trusted CA.
2. Build locally with `build-windows.bat`.
3. Sign with `sign-windows.bat` after setting:
   - `SIGN_CERT` = path to your `.pfx`
   - `SIGN_PASSWORD` = certificate password

Unsigned builds are still safe when you compile them yourself from the source package.
