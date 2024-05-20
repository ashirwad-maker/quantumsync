package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func newEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuf)
	return keyBuf
}

// The function encrypts the data from an io.Reader and write the encrypted data on the io.Writer
// using the AES encryption algorithm in CTR mode.
func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize())

	//Create a buffer for the IV of size equal to the block size(16 bytes)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}
	// Prepend the iv file
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}

	// Making the buffer to th specific size so that io.Copy can work properly.
	buf := make([]byte, 32*1024)
	stream := cipher.NewCTR(block, iv) // Creates a CTR mode stream cipher with the AES block and IV
	nw := block.BlockSize()

	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return nw, nil
}

func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	// Read the IV from the given io.Reader
	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}
	// fmt.Println(iv)

	buf := make([]byte, 32*1024)
	stream := cipher.NewCTR(block, iv)
	nw := block.BlockSize()

	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return nw, err
}
