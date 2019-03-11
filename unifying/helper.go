package unifying

import "errors"

func KeyData2Key(keydata []byte) (key []byte, err error) {
	if keydata == nil || len(keydata) != 16 {
		err = errors.New("invalid keydata")
	}

	//ToDo: keydata with nonces all 0x00 shouldn't be considered, as it is for unencrypted devices

	key = make([]byte, 16)
	key[2] = keydata[0]
	key[1] = keydata[1] ^ 0xff
	key[5] = keydata[2] ^ 0xff
	key[3] = keydata[3]
	key[14] = keydata[4]
	key[11] = keydata[5]
	key[9] = keydata[6]
	key[0] = keydata[7]
	key[8] = keydata[8]
	key[6] = keydata[9] ^ 0x55
	key[4] = keydata[10]
	key[15] = keydata[11]
	key[10] = keydata[12] ^ 0xff
	key[12] = keydata[13]
	key[7] = keydata[14]
	key[13] = keydata[15] ^ 0x55
	return
}


