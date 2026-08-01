# Verified activation links

Production activation uses one exact HTTPS origin and the exact path
`/activate`. The `belcanto://` scheme is a development-only fallback.

Publish the rendered templates at the activation origin without redirects:

- `apple-app-site-association.template.json` as
  `/.well-known/apple-app-site-association`, with `Content-Type:
  application/json`;
- `assetlinks.template.json` as `/.well-known/assetlinks.json`, with
  `Content-Type: application/json`.

Required release inputs:

- Apple Team ID;
- the same iOS bundle identifier supplied as
  `BELCANTO_IOS_BUNDLE_IDENTIFIER`;
- the same Android package supplied as `BELCANTO_ANDROID_PACKAGE`;
- every SHA-256 certificate fingerprint that can sign the production Android
  build (normally the Play App Signing certificate, plus any deliberately
  supported direct-distribution certificate);
- `EXPO_PUBLIC_ACTIVATION_ORIGIN`, matching the HTTPS host serving both files.

Replace every `${...}` placeholder during deployment and reject a rendered
file that still contains one. Verify the association from a real signed iOS
and Android build before issuing production invitations.

The API produces `https://<origin>/activate#token=<opaque-token>`. URL fragments
are handled locally by the application and are not sent in the HTTP request.
The fallback web server, CDN, analytics, error reporting, and redirect tooling
must never reconstruct, log, or persist the fragment or activation token.
