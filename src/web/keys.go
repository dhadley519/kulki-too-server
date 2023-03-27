package web

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
)

var priv *rsa.PrivateKey
var pub any

func InitPublicKey() {
	block, _ := pem.Decode([]byte(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAmaVXal9/dUkmnLf430AA
IgHj90r5sBsGuEZ47Fx9Oth7fvqUHUyWN6kxArRCQdbQs55v4W8t13zowNY9eLrR
+a647cSYna/3ptJgkSYsAeDAcStpmWukIvgFnLetpSuKoWUTA9h0jz9GDqKAnLHA
qIIXD3gNXJCCtPhXa/d8P1VNJFdji5/tdQMQ5cEuqxG6JeYSFFkGxSUEGMgm9zBI
BMVGH4+Oe4f4o7Le8UFYaMACREHdhYFuvZ4nzWI/NFSEjdqDRZiab7Wixp63dWuz
/Bb5LgVCF7RcgMJfBHamzmHR1UPKOVofjME15n29xTNJn13laMeWNhy2llWTXt6i
A+HaJ3oWNR/uKNbjBFiiZKgk8f320CF5aK8TnmuErvlosWx0kKBfpkBQ4d5ysSOw
ZyYrE/PuOsaMQATiYS2mL0DcveMRSerJG7UNBPV1jXax3pkCi6zIQtPdS2s+uHD3
7HLt+dR5dE3R/WG7XK9KRUmOWmt5lV+c66zubzhb6Jzo4T87j8d0jyDtKYhxi0xd
l00fO4CKJ5IaLQamgjvk2KCGdq6A81uLnFQ3tEu7KmrBS/npKsPvelUTcQx5TOrr
NWdMH8y42VrR5K8ZlxgfBUNJvqepONUDJqBmeySNjlC8iklpqhtij23IO2ssEYbT
pSboqBfKVFcaiJ5DVWJn4WUCAwEAAQ==
-----END PUBLIC KEY-----`))
	if block == nil {
		log.Fatal("failed to decode PEM block containing public key")
	}

	var err error
	pub, err = x509.ParsePKIXPublicKey(block.Bytes)

	switch pub.(type) {
	case *rsa.PublicKey:
		fmt.Println("pub is of type RSA")
	default:
		panic("unknown type of public key")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func InitPrivateKey() {
	block, _ := pem.Decode([]byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`))
	if block == nil {
		log.Fatal("failed to decode PEM block containing private key")
	}
	var err error
	priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
}
