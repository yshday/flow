package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

// Config holds email configuration
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Client is an email client
type Client struct {
	config Config
}

// NewClient creates a new email client
func NewClient(config Config) *Client {
	return &Client{
		config: config,
	}
}

// EmailData represents data for email templates
type EmailData struct {
	To          []string
	Subject     string
	Body        string
	TemplateStr string
	Data        interface{}
}

// Send sends an email
func (c *Client) Send(data EmailData) error {
	// Skip if SMTP is not configured
	if c.config.Host == "" || c.config.Username == "" {
		return nil
	}

	// Prepare email body
	var body string
	if data.TemplateStr != "" {
		tmpl, err := template.New("email").Parse(data.TemplateStr)
		if err != nil {
			return fmt.Errorf("failed to parse email template: %w", err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data.Data); err != nil {
			return fmt.Errorf("failed to execute email template: %w", err)
		}
		body = buf.String()
	} else {
		body = data.Body
	}

	// Prepare message
	msg := fmt.Sprintf("From: %s\r\n", c.config.From)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(data.To, ","))
	msg += fmt.Sprintf("Subject: %s\r\n", data.Subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=UTF-8\r\n"
	msg += "\r\n"
	msg += body

	// Setup authentication
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)

	// Send email
	addr := fmt.Sprintf("%s:%s", c.config.Host, c.config.Port)
	err := smtp.SendMail(addr, auth, c.config.From, data.To, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendIssueAssigned sends an email when an issue is assigned
func (c *Client) SendIssueAssigned(to, issueKey, issueTitle, assignedBy, projectName string) error {
	template := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .issue-key { font-weight: bold; color: #4CAF50; }
        .footer { margin-top: 20px; padding: 10px; font-size: 12px; color: #666; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>이슈가 배정되었습니다</h2>
        </div>
        <div class="content">
            <p>안녕하세요,</p>
            <p><strong>{{.AssignedBy}}</strong>님이 당신에게 이슈를 배정했습니다.</p>
            <p>
                <span class="issue-key">{{.IssueKey}}</span>: {{.IssueTitle}}
            </p>
            <p>프로젝트: <strong>{{.ProjectName}}</strong></p>
        </div>
        <div class="footer">
            <p>이 메일은 자동으로 발송되었습니다.</p>
        </div>
    </div>
</body>
</html>
`

	return c.Send(EmailData{
		To:          []string{to},
		Subject:     fmt.Sprintf("[%s] 이슈가 배정되었습니다", issueKey),
		TemplateStr: template,
		Data: map[string]string{
			"IssueKey":    issueKey,
			"IssueTitle":  issueTitle,
			"AssignedBy":  assignedBy,
			"ProjectName": projectName,
		},
	})
}

// SendCommentAdded sends an email when a comment is added to an issue
func (c *Client) SendCommentAdded(to, issueKey, issueTitle, commenter, commentText, projectName string) error {
	template := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .issue-key { font-weight: bold; color: #2196F3; }
        .comment { background-color: white; padding: 15px; margin: 10px 0; border-left: 3px solid #2196F3; }
        .footer { margin-top: 20px; padding: 10px; font-size: 12px; color: #666; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>새 댓글이 추가되었습니다</h2>
        </div>
        <div class="content">
            <p>안녕하세요,</p>
            <p><strong>{{.Commenter}}</strong>님이 이슈에 댓글을 남겼습니다.</p>
            <p>
                <span class="issue-key">{{.IssueKey}}</span>: {{.IssueTitle}}
            </p>
            <div class="comment">
                {{.CommentText}}
            </div>
            <p>프로젝트: <strong>{{.ProjectName}}</strong></p>
        </div>
        <div class="footer">
            <p>이 메일은 자동으로 발송되었습니다.</p>
        </div>
    </div>
</body>
</html>
`

	return c.Send(EmailData{
		To:          []string{to},
		Subject:     fmt.Sprintf("[%s] 새 댓글이 추가되었습니다", issueKey),
		TemplateStr: template,
		Data: map[string]string{
			"IssueKey":    issueKey,
			"IssueTitle":  issueTitle,
			"Commenter":   commenter,
			"CommentText": commentText,
			"ProjectName": projectName,
		},
	})
}
