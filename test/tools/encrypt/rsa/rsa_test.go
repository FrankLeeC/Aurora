package test

import (
	"testing"

	"github.com/FrankLeeC/Aurora/tools/encrypt/rsa"
)

const s = "hello 世界"

func TestPKCS1FromFile(t *testing.T) {
	pubKey, err := rsa.GeneratePublicKeyFromFile("./pkcs1/rsa_public.pem")
	if err != nil {
		t.Errorf("generate public key from file error: %s\n", err.Error())
		return
	}
	priKey, err := rsa.GeneratePrivateKeyFromFile("./pkcs1/rsa_private.pem", rsa.PKCS1)
	if err != nil {
		t.Errorf("generate private key from file error: %s\n", err.Error())
		return
	}
	cipherBytes, err := rsa.Encrypt(s, pubKey)
	if err != nil {
		t.Errorf("encrypt error: %s\n", err.Error())
		return
	}
	plainBytes, err := rsa.Decrypt(string(cipherBytes), priKey)
	if err != nil {
		t.Errorf("decrypt error: %s\n", err.Error())
		return
	}
	t.Log(string(plainBytes))
	if string(plainBytes) == s {
		t.Log("pass!")
	} else {
		t.Log("fail!")
	}

}

func TestPKCS1FromString(t *testing.T) {
	pubKey, err := rsa.GeneratePublicKey([]byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCwHsMGYrbgR1XgkH0AsrZQyAhZ
oXluYNFpZfC6TE33ykJrNge+9LwP9Hs0tO7FcY52iGgxZ0Fa8JcSfRzBlJnJz14k
KGlpLEcnPBhV3W77s4RvM+ZKSEv57Z2HPRgXgJue/mcKI+DzjYGpfhMKa5+JG2Zm
Sw6yQZ7BFW4S9vGZDwIDAQAB
-----END PUBLIC KEY-----		
`))
	if err != nil {
		t.Errorf("generate public key error: %s\n", err.Error())
		return
	}
	priKey, err := rsa.GeneratePrivateKey([]byte(`
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCwHsMGYrbgR1XgkH0AsrZQyAhZoXluYNFpZfC6TE33ykJrNge+
9LwP9Hs0tO7FcY52iGgxZ0Fa8JcSfRzBlJnJz14kKGlpLEcnPBhV3W77s4RvM+ZK
SEv57Z2HPRgXgJue/mcKI+DzjYGpfhMKa5+JG2ZmSw6yQZ7BFW4S9vGZDwIDAQAB
AoGAFD201EsMEYKhUAnLFAV2Bpq2uvZf6lueNarNm9uhKlVIhOHUyhF+e6bxcgFJ
X8/JO745m1PuovD7q9hNMcJOWj51VHalfXClNiUKavIhiUGmw/H7RS3q8rLCMaiN
pJjANPK6leMySVplD27rk8r1Sfj9gFw+lOK+MkkwlAnsjeECQQDY9lCsLcMFYKfR
op7QJHoAGgwtFsH4l1AgrP6TOl91nOfLZPAM6rMTBegHWKXjRqGZJzYOJZLU/Acn
Opg6/KBTAkEAz88v7y7d0fm1UBDWkw18BLcCnD/+jOFe997Ygrw8Pn50NeELWJ8W
T03gmQWeXM4zSxFThsTrt5O8iNGNI4h81QJBAIpMIEpQdULFRNQFM/R7e+T6tY48
NnKuiR37B99zUwwWc06cTcP+Cx3yIuAj6sI/8Jw+eV91Je5rpGemGwlmpQ8CQQDG
yPeegiN4r7iMXX8U7io7TGF33BOA6jlxPL+515x9X3OE8sBqxsuNkv6NAn3pYupY
HbvbyFV/pxgLfQDZA7/9AkAr6q42lykYZaipxS1gAqo3DFnLoA3SFoPme3XiAGqr
fRiCujrulI4c+s0Vk6NpW3TFkhTvkx+P6j9MK0MRfDVH
-----END RSA PRIVATE KEY-----		
		`), rsa.PKCS1)
	if err != nil {
		t.Errorf("generate private key error: %s\n", err.Error())
		return
	}
	cipherText, err := rsa.Encrypt(s, pubKey)
	if err != nil {
		t.Errorf("encrypt error: %s\n", err.Error())
		return
	}

	plainBytes, err := rsa.Decrypt(string(cipherText), priKey)
	if err != nil {
		t.Errorf("decrypt error: %s\n", err.Error())
		return
	}
	t.Log(string(plainBytes))
	if string(plainBytes) == s {
		t.Log("pass!")
	} else {
		t.Log("fail!")
	}
}

func TestPKCS8FromFile(t *testing.T) {
	pubKey, err := rsa.GeneratePublicKeyFromFile("./pkcs8/rsa_public.pem")
	if err != nil {
		t.Errorf("generate public key from file error: %s\n", err.Error())
		return
	}
	priKey, err := rsa.GeneratePrivateKeyFromFile("./pkcs8/rsa_private_pkcs8.pem", rsa.PKCS8)
	if err != nil {
		t.Errorf("generate private key from file error: %s\n", err.Error())
		return
	}
	cipherBytes, err := rsa.Encrypt(s, pubKey)
	if err != nil {
		t.Errorf("encrypt error: %s\n", err.Error())
		return
	}
	plainBytes, err := rsa.Decrypt(string(cipherBytes), priKey)
	if err != nil {
		t.Errorf("decrypt error: %s\n", err.Error())
		return
	}
	t.Log(string(plainBytes))
	if string(plainBytes) == s {
		t.Log("pass!")
	} else {
		t.Log("fail!")
	}
}

func TestPKCS8FromString(t *testing.T) {
	pubKey, err := rsa.GeneratePublicKey([]byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCwHsMGYrbgR1XgkH0AsrZQyAhZ
oXluYNFpZfC6TE33ykJrNge+9LwP9Hs0tO7FcY52iGgxZ0Fa8JcSfRzBlJnJz14k
KGlpLEcnPBhV3W77s4RvM+ZKSEv57Z2HPRgXgJue/mcKI+DzjYGpfhMKa5+JG2Zm
Sw6yQZ7BFW4S9vGZDwIDAQAB
-----END PUBLIC KEY-----		
		`))
	if err != nil {
		t.Errorf("generate public key error: %s\n", err.Error())
		return
	}
	priKey, err := rsa.GeneratePrivateKey([]byte(`
-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBALAewwZituBHVeCQ
fQCytlDICFmheW5g0Wll8LpMTffKQms2B770vA/0ezS07sVxjnaIaDFnQVrwlxJ9
HMGUmcnPXiQoaWksRyc8GFXdbvuzhG8z5kpIS/ntnYc9GBeAm57+Zwoj4PONgal+
Ewprn4kbZmZLDrJBnsEVbhL28ZkPAgMBAAECgYAUPbTUSwwRgqFQCcsUBXYGmra6
9l/qW541qs2b26EqVUiE4dTKEX57pvFyAUlfz8k7vjmbU+6i8Pur2E0xwk5aPnVU
dqV9cKU2JQpq8iGJQabD8ftFLeryssIxqI2kmMA08rqV4zJJWmUPbuuTyvVJ+P2A
XD6U4r4ySTCUCeyN4QJBANj2UKwtwwVgp9GintAkegAaDC0WwfiXUCCs/pM6X3Wc
58tk8AzqsxMF6AdYpeNGoZknNg4lktT8Byc6mDr8oFMCQQDPzy/vLt3R+bVQENaT
DXwEtwKcP/6M4V733tiCvDw+fnQ14QtYnxZPTeCZBZ5czjNLEVOGxOu3k7yI0Y0j
iHzVAkEAikwgSlB1QsVE1AUz9Ht75Pq1jjw2cq6JHfsH33NTDBZzTpxNw/4LHfIi
4CPqwj/wnD55X3Ul7mukZ6YbCWalDwJBAMbI956CI3ivuIxdfxTuKjtMYXfcE4Dq
OXE8v7nXnH1fc4TywGrGy42S/o0Cfeli6lgdu9vIVX+nGAt9ANkDv/0CQCvqrjaX
KRhlqKnFLWACqjcMWcugDdIWg+Z7deIAaqt9GIK6Ou6Ujhz6zRWTo2lbdMWSFO+T
H4/qP0wrQxF8NUc=
-----END PRIVATE KEY-----			
				`), rsa.PKCS8)
	if err != nil {
		t.Errorf("generate private key error: %s\n", err.Error())
		return
	}
	cipherText, err := rsa.Encrypt(s, pubKey)
	if err != nil {
		t.Errorf("encrypt error: %s\n", err.Error())
		return
	}

	plainBytes, err := rsa.Decrypt(string(cipherText), priKey)
	if err != nil {
		t.Errorf("decrypt error: %s\n", err.Error())
		return
	}
	t.Log(string(plainBytes))
	if string(plainBytes) == s {
		t.Log("pass!")
	} else {
		t.Log("fail!")
	}
}
