package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailConfig SMTP 邮件配置
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

// EmailSender 邮件发送器
type EmailSender struct {
	config *EmailConfig
}

// NewEmailSender 创建邮件发送器
func NewEmailSender(config *EmailConfig) *EmailSender {
	return &EmailSender{
		config: config,
	}
}

// SendEmail 发送邮件
func (s *EmailSender) SendEmail(to, subject, body string) error {
	if s.config == nil {
		return fmt.Errorf("email config is nil")
	}

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = s.config.From
	headers["To"] = to
	headers["Subject"] = subject
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 组装邮件内容
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// SMTP 认证
	auth := smtp.PlainAuth(
		"",
		s.config.Username,
		s.config.Password,
		s.config.Host,
	)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	if s.config.UseTLS {
		return s.sendEmailWithTLS(addr, auth, to, []byte(message))
	}

	return smtp.SendMail(
		addr,
		auth,
		s.config.From,
		[]string{to},
		[]byte(message),
	)
}

// sendEmailWithTLS 使用 TLS 发送邮件
func (s *EmailSender) sendEmailWithTLS(addr string, auth smtp.Auth, to string, msg []byte) error {
	// 解析主机名（用于 TLS 验证）
	host := strings.Split(addr, ":")[0]

	// 连接到 SMTP 服务器
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	// STARTTLS
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false, // 生产环境应验证证书
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// 认证
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// 设置发件人
	if err = client.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// 设置收件人
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}
