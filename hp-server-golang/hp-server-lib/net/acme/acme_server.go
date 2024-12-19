package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

var ConfigAcme *AcmeServer

type AcmeServer struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
	client       *lego.Client
}

func (u *AcmeServer) GetEmail() string {
	return u.Email
}
func (u AcmeServer) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *AcmeServer) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
func (u *AcmeServer) GenCert(domain string) (*certificate.Resource, error) {
	request := certificate.ObtainRequest{
		Domains: []string{domain}, // 这里如果有多个，就写多个就好了，可以是多个域名
		Bundle:  true,             // 这里如果是true，将把颁发者证书一起返回，也就是返回里面certificates.IssuerCertificate
	}
	// 开始申请证书
	certificates, err := u.client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}
	// 申请完了后，里面会带有证书的PrivateKey Certificate，都为[]byte格式，需要存储的自行转为string即可
	return certificates, nil
}

func StartAcmeServer(email, httpPort string) error {
	// 创建myUser用户对象。新对象需要email和私钥才能启动，私钥需自己生成
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	acmeSer := &AcmeServer{
		Email: email,
		key:   privateKey,
	}
	config := lego.NewConfig(acmeSer)
	// 此处配置密钥的类型和密钥申请的地址，记得上线后替换成 lego.LEDirectoryProduction ，测试环境下就用 lego.LEDirectoryStaging
	config.CADirURL = lego.LEDirectoryProduction
	config.Certificate.KeyType = certcrypto.RSA2048
	// 创建一个client与CA服务器通信
	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}
	// 此处需要进行申请证书的chanlldge，必须监听80和443端口，这样才能让Let's Encrypt访问到我们的服务器
	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", httpPort))
	if err != nil {
		return err
	}
	// 把这个客户端注册，传递给myUser用户里
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	acmeSer.client = client
	acmeSer.Registration = reg
	ConfigAcme = acmeSer
	return nil
}
