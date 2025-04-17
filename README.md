# Crypt 

A library for Steganography workflows. AES, QRCode, DCT embedding and extraction.

> Steganography hides presence; once the carrier is suspect, AES does the confidentiality.

*Threat model*: be **secretive**. Then, if found out, still be safe. Live through extreme compression.

## Example workflow:

``` mermaid
flowchart TB
    %% ── 1st ROW ────────────────────────────────────────
    subgraph HIDE["HIDE  (crypt encrypt … embed)"]
        direction LR
        plaintext["Plain text"] -->|AES‑256‑GCM| cipher
        cipher -->|base64| b64
        b64 -->|"QR encode"| qr
        qr -->|"binary bits"| bits
        bits -->|"DCT embed"| stego
    end

    %% ── 2nd ROW ────────────────────────────────────────
    subgraph REVEAL["REVEAL  (crypt decrypt)"]
        direction LR
        extract["Extract DCT"] -->|"QR decode"| qr2
        qr2 -->|base64| b642
        b642 -->|AES‑256‑GCM| plain2
        plain2["Plain text"]
    end

    %% ── CROSS‑ROW LINK ─────────────────────────────────
    stego --> extract

    %% optional styling for the carrier JPEG
    classDef carrier fill:#ffd5b3,stroke:#e49a46,stroke-width:1.5px,color:#000;
    class stego carrier;
```

``` bash
crypt encrypt text q mysecurepassword qrcode binary embed "./test/input.jpeg" test/out_embedded.jpeg
```

``` bash
crypt decrypt image ./test/out_embedded.jpeg extract text mysecurepassword
```

See the embedded `sxiv ./test/out_embedded.jpeg`. The secret is there and no image distortions!

## Safezone

- Payload size ≈ 2 kB: QR Version 5‑L at 65 % `JPEG` quality is an empirical safezone. 
- E.g., will be there after linkedin compresses your image in a tiny `jpeg`, for example.
- No authenticity yet: GCM gives you integrity only if the key is secret. If you switch to a stronger KDF, you’re fine; otherwise add an HMAC.
