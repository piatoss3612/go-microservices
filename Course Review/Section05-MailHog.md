# Section05 Review

### 1. docker-compose로 MailHog 실행

- MailHog: Email(SMTP) testing tool
- [MailHog Github](https://github.com/mailhog/MailHog)

```yaml
version: '3'

services:

  mailhog:
    image: 'mailhog/mailhog:latest'
    ports:
      - "1025:1025" # smtp server
      - "8025:8025" # web user interface
    
```

### 2. MailHog를 사용한 Email 전송 테스트

```cmd
go get github.com/vanng822/go-premailer/premailer
go get github.com/xhit/go-simple-mail/v2
```

```go
package main

import (
	"bytes"
	"html/template"
	"os"
	"strconv"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)
```

#### Email 형식

```go
type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}
```

#### Email 전송을 담당하는 Mail 객체 및 팩토리 함수

```go
type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

func createMail() Mail {
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		port = 1025
	}
	return Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
	}
}
```

#### SMTP 메시지(이메일) 전송

```go
func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

    // html/template 패키지를 사용해 메시지를 변환하기 위해 키:값 형태의 map을 사용
	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data


    // 1. msg를 HTML 형식으로 변환
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

    // 2. msg를 평문 형식으로 변환
	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

    // 3. SMTP 서버 설정 초기화
	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

    // 4. SMTP 클라이언트 연결
	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	// 5. 이메일 생성
	email := mail.NewMSG() 
	email.SetFrom(msg.From). // 보내는 이 설정
		AddTo(msg.To). // 받는 이 설정
		SetSubject(msg.Subject). // 제목 설정
		SetBody(mail.TextPlain, plainMessage). // 본문을 평문으로된 메시지로 설정
		AddAlternative(mail.TextHTML, formattedMessage) // 본문이 제대로 표시되지 않을 경우를 대비해 HTML 버전의 메시지를 설정

    // 첨부파일 추가
	if len(msg.Attachments) > 0 {
		for _, attm := range msg.Attachments {
			email.AddAttachment(attm)
		}
	}

    // 6. SMTP 클라이언트를 통해 이메일 전송
	if err = email.Send(smtpClient); err != nil {
		return err
	}

	return nil
}
```

#### 템플릿을 사용해 Message 객체를 HTML 형식으로 변환

```go
func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := "./templates/mail.html.gohtml"

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()

	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}
```

#### HTML 메일 인라인 스타일링

```go
func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	pm, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := pm.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
```

#### 템플릿을 사용해 Message 객체를 평문으로 변환

```go
func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := "./templates/mail.plain.gohtml"

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}
```

#### SMTP 통신 암호화 방식 설정

```go
func (m *Mail) getEncryption(s string) mail.Encryption {
	switch s {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSSLTLS
	}
}
```