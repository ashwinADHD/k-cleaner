# Code Signing and Notarization Guide

K-Cleaner is distributed as an unsigned binary by default. macOS Gatekeeper will block first launch until you approve the app. This guide covers signing and notarizing for public distribution.

---

## Prerequisites

1. **Apple Developer Program** membership ($99/year)
2. **Developer ID Application** certificate installed in Keychain
3. **Developer ID Installer** certificate (optional, for PKG)
4. Xcode Command Line Tools: `xcode-select --install`

Verify certificates:

```bash
security find-identity -v -p codesigning
```

Look for: `Developer ID Application: Your Name (TEAMID)`

---

## Sign the CLI binary

```bash
cd kcleaner
make build

codesign --force --options runtime --sign "Developer ID Application: Your Name (TEAMID)" \
  --timestamp \
  ./kclean

codesign --verify --verbose ./kclean
spctl --assess --verbose ./kclean
```

---

## Sign K-Cleaner.app

Build the app bundle first:

```bash
make app
```

Sign the embedded binary, then the bundle:

```bash
APP="K-Cleaner.app"
IDENTITY="Developer ID Application: Your Name (TEAMID)"

codesign --force --options runtime --sign "$IDENTITY" --timestamp \
  "$APP/Contents/MacOS/kclean"

codesign --force --options runtime --sign "$IDENTITY" --timestamp \
  --entitlements entitlements.plist \
  "$APP"

codesign --verify --deep --strict --verbose=2 "$APP"
```

### Sample entitlements (`entitlements.plist`)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.cs.disable-library-validation</key>
    <false/>
</dict>
</plist>
```

K-Cleaner does not require sandbox entitlements. Full Disk Access is granted by the user in System Settings.

---

## Notarize for Gatekeeper

1. Create a zip of the app:

```bash
ditto -c -k --keepParent "K-Cleaner.app" "K-Cleaner.zip"
```

2. Submit for notarization:

```bash
xcrun notarytool submit "K-Cleaner.zip" \
  --apple-id "your@email.com" \
  --team-id "TEAMID" \
  --password "@keychain:AC_PASSWORD" \
  --wait
```

3. Staple the ticket:

```bash
xcrun stapler staple "K-Cleaner.app"
spctl --assess --verbose "K-Cleaner.app"
```

---

## Hardened Runtime flags

Always sign with `--options runtime` for notarization compatibility.

If you add helper tools or plugins later, each nested binary must be signed individually before signing the outer bundle.

---

## Full Disk Access after signing

Signing does **not** grant Full Disk Access. Users must still:

1. System Settings → Privacy & Security → Full Disk Access
2. Add **K-Cleaner** or **kclean**
3. Restart the app

Run `kclean permissions` to verify access.

---

## CI signing (optional)

Store certificates and notarization credentials as GitHub Actions secrets:

- `BUILD_CERTIFICATE_BASE64`
- `P12_PASSWORD`
- `KEYCHAIN_PASSWORD`
- `NOTARIZE_APPLE_ID`
- `NOTARIZE_TEAM_ID`
- `NOTARIZE_PASSWORD`

Use `security import` in CI to load the certificate, then run the sign + notarytool steps above.

---

## Unsigned distribution (current default)

For personal use or development:

1. Build: `make app`
2. Copy to Applications: `cp -R K-Cleaner.app /Applications/`
3. First launch: **Right-click → Open → Open**

No Apple Developer account required.
