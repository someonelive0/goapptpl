package utils

import "crypto/tls"

/*
	TLS and HTTPS cert, self signed cert and key pem file with PKCS#1-PEM

openssl genrsa -out key.pem 2048
openssl rsa -in key.pem -pubout -out pub.pem
openssl req -new -x509 -sha256 -key key.pem -out cert.pem -days 36500 -subj "/C=CN/ST=BJ/L=BJ/O=idss/OU=dsmc/CN=dsmc.idss.com"
openssl rsa -in key.pem -noout -text
openssl rsa -pubin -in pub.pem -noout -text
openssl x509 -in cert.pem -noout -text
openssl rsautl -encrypt -inkey pub.pem -pubin -in message.txt -out enc.dat
openssl rsautl -decrypt -inkey key.pem -in enc.dat -out new_message.txt
*/
func TLSConfig() *tls.Config {
	//cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	cert, err := tls.X509KeyPair(CertPem, KeyPem)
	if err != nil {
		return nil
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return cfg
}

var (
	CertPem = []byte(`-----BEGIN CERTIFICATE-----
MIIDjzCCAnegAwIBAgIJAMef0WZSrVTXMA0GCSqGSIb3DQEBCwUAMF0xCzAJBgNV
BAYTAkNOMQswCQYDVQQIDAJCSjELMAkGA1UEBwwCQkoxDTALBgNVBAoMBGlkc3Mx
DTALBgNVBAsMBGRzbWMxFjAUBgNVBAMMDWRzbWMuaWRzcy5jb20wIBcNMjQxMDI0
MDY0MDA5WhgPMjEyNDA5MzAwNjQwMDlaMF0xCzAJBgNVBAYTAkNOMQswCQYDVQQI
DAJCSjELMAkGA1UEBwwCQkoxDTALBgNVBAoMBGlkc3MxDTALBgNVBAsMBGRzbWMx
FjAUBgNVBAMMDWRzbWMuaWRzcy5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCmTBzRNySgQoWNVMx04fF4nrmTyjrjU3V6flys2XyNGVKoSBetpaXb
tN8nsfAX5LAYgFBHjNkU+6VIgk8kUwXr+gGBn6/5J0PpBpz+nirlL0WUTe56oCvR
0caPEjctjd2mi8CxvROiPq5ShXmLfUJvk2iaqcU6PaXbMeM8rpenqhk3lglnfwpt
PMIdzkAKL+xEeRlroZzOOd4QBGflUuW8e5ceI4p8tuofmxiPDEsg1bYZgCGWTWsG
mW/6kIZdYcGwU7vXbqYq8HvFyvIEgdxlr8SORcTYE9FVuilaXOidTjsVMoCsgtua
WY0Kj3h2RUCmeAGIeUkt6vZUxHp0GHjzAgMBAAGjUDBOMB0GA1UdDgQWBBRZxWmN
PoNJw5X7wL0b2ieszVHA9TAfBgNVHSMEGDAWgBRZxWmNPoNJw5X7wL0b2ieszVHA
9TAMBgNVHRMEBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQApecrWDNUYEFC+Cm2q
OIaI+OuKor9A3RXHJ/7aTZtFVSwj3wpcnKn1+nW/qExMhqVQ6cPHJywD5TGfvq0c
CaIL6mTCQx5c2W1KEh7DZSePfj/knJQFqlKvpZRM072FTDrouqy+ptIFpJvvA1sE
ON1Htf6Zu6az4Ta9DRaU/RyfkYbqKG4lmgkbielLfjzKxsGps7hrdOd6Pen1duBQ
3y6EMoit64DS0uoPIeQNxljxUWXgwgAE1Wp+gKNKhp/x2OO/vpT3C9vwUZDUFVYY
B5eY1aV7xSZ/Ngevv+aLcW7wiQ9o1NTgOyA2ECO50kSZxvn1H0OanZeGJfjVFj+G
zT3i
-----END CERTIFICATE-----`)

	KeyPem = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEApkwc0TckoEKFjVTMdOHxeJ65k8o641N1en5crNl8jRlSqEgX
raWl27TfJ7HwF+SwGIBQR4zZFPulSIJPJFMF6/oBgZ+v+SdD6Qac/p4q5S9FlE3u
eqAr0dHGjxI3LY3dpovAsb0Toj6uUoV5i31Cb5NomqnFOj2l2zHjPK6Xp6oZN5YJ
Z38KbTzCHc5ACi/sRHkZa6GczjneEARn5VLlvHuXHiOKfLbqH5sYjwxLINW2GYAh
lk1rBplv+pCGXWHBsFO7126mKvB7xcryBIHcZa/EjkXE2BPRVbopWlzonU47FTKA
rILbmlmNCo94dkVApngBiHlJLer2VMR6dBh48wIDAQABAoIBADpOQWRRUzQlXrH4
416vwXwrGqHoq9D8eBokp9Wqw3KtSD/cVwD5LTflKMDAMJAQDHXqtzr+9TXYLVFI
7vqraU4db93E4WwYCkVvniffyOZmNp0S3eR8uCBuzpFnL5V3QhmcskkfI/0kwPYJ
+l2f42U/Z9OiZpZO+xHpYeTqyh1sg49OuLsc/2Ob9OiC6Sy9peI+ki4R0ChIaS3n
y5uY3nBaRdgTPU5J9cs8ruNzG66Yt6V6bgyaXAdvWoFKjoT3b6ah7I0hsX5iGwTT
+Vip5+MoI4JqRD6edEuh5Z1OqTwtC1l1te/rDdnK9EVy3XDYEjPEivOepHmRPRcP
DuZANNECgYEAzuOEtXukaCd8bw/sY/rAJPKDnFRF57Wn7iGqxZh9loL6n8dcosTV
BZYZwHTBEYT/aFOAjyfAihnzSBlLDAp2Db3mCEFA6QMjHzSFUF0VOKc5SeNqXvs9
eKosWMQdAD+ib8bPL0hTBvEhxdWWTvFnSEHnGPiaSSDFKxB+0uaKJM0CgYEAzcXh
xvdPYw0YT54YNZgP0xYamySMovULwonp2YWrdo+0ouhsbJHmQlTvryKsB0yCEw7F
ZgMOdvmUYBGF1LSsU1SnMtnHs6mT/duYI/P+iYc/TRJnonXXXHUG/mK66+v2RDna
sLf6pU1Lf0aKkNhAn8dEj8m4XKGpXOwIciDDFL8CgYEAsGZvtenpYWEhiPTTwt9/
S0F4FCgKvqk1uSX9nKMLmfStyuRKSQJ4+11jMaSbJdv3hbWE7Qqg8V90/mmKgoa8
57Sd2TYCKWsiXC4E6WOkf3ydrTF5dejUHflC/KCidZ7MWm/yIceR+15IRI17rm3I
eWSvravyqR2G39QdvqcQ7JUCgYB4rzc0/3VDDboVcA6Y2D9nuQ4Pscb+CCRGi6Zo
mlou5ie2aAS3RHa8rp4IpJgqi7e6P66MnvxL0SMxmPVaBEERepO5Yjsa5zlR6Qn5
BDBkLrt0k3fOs7iElGpupi8lETZVW20kujK54nSGCDRasUptq2xNvKxxP6taQWDO
tuJTdwKBgA6e5hMniDRztnwiMM1HtdHPxRCiZ5f2eYNQV5LWNz8yFdDqEjnvDk0a
2qJkFRyAEF4Eg8nPzZz22r37S7yjbNxVY1lcnntVedHLH7shJAHKOHnL9dXcpPre
PVqbtRHaM2bhwBhhI9QjPbRs8r0VSm8h9JAZifvkpdg4M7jixy8d
-----END RSA PRIVATE KEY-----`)

	PubPem = []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApkwc0TckoEKFjVTMdOHx
eJ65k8o641N1en5crNl8jRlSqEgXraWl27TfJ7HwF+SwGIBQR4zZFPulSIJPJFMF
6/oBgZ+v+SdD6Qac/p4q5S9FlE3ueqAr0dHGjxI3LY3dpovAsb0Toj6uUoV5i31C
b5NomqnFOj2l2zHjPK6Xp6oZN5YJZ38KbTzCHc5ACi/sRHkZa6GczjneEARn5VLl
vHuXHiOKfLbqH5sYjwxLINW2GYAhlk1rBplv+pCGXWHBsFO7126mKvB7xcryBIHc
Za/EjkXE2BPRVbopWlzonU47FTKArILbmlmNCo94dkVApngBiHlJLer2VMR6dBh4
8wIDAQAB
-----END PUBLIC KEY-----`)
)
