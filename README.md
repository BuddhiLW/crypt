# Crypt 

A library for Steganography workflows. AES, QRCode, DCT embedding and extraction.

> Steganography hides presence; once the carrier is suspect, AES does the confidentiality.

*Threat model*: be **secretive**. Then, if found out, still be safe. Live through extreme compression.

## Example workflow:

``` mermaid
flowchart LR
    %% ─── shared node so both subgraphs can see it ─────────
    stego((Stego JPEG)):::carrier

    %% ─── HIDE PIPELINE ────────────────────────────────────
    subgraph "HIDE (crypt encrypt … embed)"
        plaintext["Plain text"] -->|AES‑256‑GCM| cipher
        cipher["Cipher‑text"] -->|base64| b64
        b64 -->|"QR encode"| qr
        qr["QR PNG"] -->|"binary bits"| bits
        bits -->|"DCT embed"| stego
    end

    %% ─── REVEAL PIPELINE ──────────────────────────────────
    subgraph "REVEAL (crypt decrypt)"
        stego -->|"Extract DCT"| bits2
        bits2 -->|"QR decode"| qr2
        qr2["QR PNG"] -->|base64| b642
        b642 -->|AES‑256‑GCM| plain2
        plain2["Plain text"]
    end

    %% ─── OPTIONAL STYLES ──────────────────────────────────
    classDef carrier fill:#ffd5b3,stroke:#e49a46,stroke-width:1.5px,color:#000;
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
