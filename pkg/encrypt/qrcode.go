package encrypt

import qrcode "github.com/skip2/go-qrcode"

func CreateQRCodeBytes(data string) ([]byte, error) {
	return qrcode.Encode(data, qrcode.Highest, 256)
}
