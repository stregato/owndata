package fs

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

type encryptingReadSeeker struct {
	inputSeeker io.ReadSeeker
	cipher      cipher.Stream
}

func (er *encryptingReadSeeker) Read(p []byte) (n int, err error) {
	n, err = er.inputSeeker.Read(p)
	if n > 0 {
		er.cipher.XORKeyStream(p[:n], p[:n])
	}
	return n, err
}

func (er *encryptingReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return er.inputSeeker.Seek(offset, whence)
}

func encryptReader(inputSeeker io.ReadSeeker, key []byte, iv []byte) (io.ReadSeeker, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)

	return &encryptingReadSeeker{
		inputSeeker: inputSeeker,
		cipher:      stream,
	}, nil
}

type decryptingWriter struct {
	outputWriter io.Writer
	cipher       cipher.Stream
}

func (dw *decryptingWriter) Write(p []byte) (n int, err error) {
	dw.cipher.XORKeyStream(p, p)
	return dw.outputWriter.Write(p)
}

func decryptWriter(outputWriter io.Writer, key []byte, iv []byte) (io.Writer, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)

	return &decryptingWriter{
		outputWriter: outputWriter,
		cipher:       stream,
	}, nil
}
