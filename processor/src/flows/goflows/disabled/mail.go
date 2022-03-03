// mail.go	functions for GoFlows to send mail

package goflows

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strings"
)

var MailFuncs = map[string]interface{}{
	"SendMail":               SendMail,
	"SendMailWithAttachment": SendMailWithAttachment,
}

// SendMail will send an email using template
func SendMail(tmplFile string, data map[string]interface{}) (string, error) {

	// check for the mail template file
	t, err := template.ParseFiles("mailTemplates/" + tmplFile)
	if err != nil {
		return "", err
	}

	msgHeader := ""
	toList := ""

	if (data["To"] != nil) && (data["To"] != "") && (data["To"] != KEYNOTFOUND) {
		msgHeader = fmt.Sprintf("To: %v\r\n", data["To"])
		toList = fmt.Sprintf("%v", data["To"])
	}

	if (data["Cc"] != nil) && (data["Cc"] != "") && (data["Cc"] != KEYNOTFOUND) {
		msgHeader = msgHeader + fmt.Sprintf("Cc: %v\r\n", data["Cc"])
		if toList == "" {
			toList = fmt.Sprintf("%v", data["Cc"])
		} else {
			toList = fmt.Sprintf("%v; %v", toList, data["Cc"])
		}
	}

	// TODO: Bcc
	// Sending "Bcc" messages is accomplished by including an email address in the toList,
	// but not in the "To:" or "Cc:" of msgHeader

	// sender
	sender := "goflows@connectria.com"
	if data["Sender"] != nil {
		sender = fmt.Sprintf("%v", data["Sender"])
	}

	if len(fmt.Sprintf("%v", data["Subject"])) > 0 {
		msgHeader = msgHeader + fmt.Sprintf("Subject: %v\r\n", data["Subject"])
	}

	var body bytes.Buffer
	mimeHeaders := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body.Write([]byte(fmt.Sprintf("%v%s", msgHeader, mimeHeaders)))

	t.Execute(&body, data)
	err = smtp.SendMail(os.Getenv("SMTP_SERVER"), nil, sender, strings.Split(toList, "; "), body.Bytes())
	if err != nil {
		return "", err
	}

	return "Mail sent.", nil
}

// NOTE: need to include "AttachmentFileName" and "AttachmentFilePath"
// SendMailWithAttachment will send an email using template
func SendMailWithAttachment(tmplFile string, data map[string]interface{}) (string, error) {

	// check for the mail template file
	t, err := template.ParseFiles("mailTemplates/" + tmplFile)
	if err != nil {
		return "", err
	}

	msgHeader := ""
	toList := ""

	if (data["To"] != nil) && (data["To"] != "") && (data["To"] != KEYNOTFOUND) {
		msgHeader = fmt.Sprintf("To: %v\r\n", data["To"])
		toList = fmt.Sprintf("%v", data["To"])
	}

	if (data["Cc"] != nil) && (data["Cc"] != "") && (data["Cc"] != KEYNOTFOUND) {
		msgHeader = msgHeader + fmt.Sprintf("Cc: %v\r\n", data["Cc"])
		if toList == "" {
			toList = fmt.Sprintf("%v", data["Cc"])
		} else {
			toList = fmt.Sprintf("%v; %v", toList, data["Cc"])
		}
	}

	// sender
	// TODO: Remove goflows@connectria.com as sender
	sender := "goflows@connectria.com"
	if data["Sender"] != nil {
		sender = fmt.Sprintf("%v", data["Sender"])
	}

	// TODO: Bcc
	// Sending "Bcc" messages is accomplished by including an email address in the toList,
	// but not in the "To:" or "Cc:" of msgHeader

	if len(fmt.Sprintf("%v", data["Subject"])) > 0 {
		msgHeader = msgHeader + fmt.Sprintf("Subject: %v\r\n", data["Subject"])
	}

	// headers
	delimeter := "**=myohmygibberish1234567"
	mimeHeaders := "MIME-Version: 1.0\r\n"
	mimeHeaders += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimeter)
	mimeHeaders += fmt.Sprintf("\r\n--%s\r\n", delimeter)
	mimeHeaders += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	mimeHeaders += "Content-Transfer-Encoding: 7bit\r\n"

	var body bytes.Buffer
	body.Write([]byte(fmt.Sprintf("%s%s", msgHeader, mimeHeaders)))

	// apply the template then send email
	t.Execute(&body, data)

	// attachment?
	if data["AttachmentFileName"] != nil {

		msgAttachment := fmt.Sprintf("\r\n--%s\r\n", delimeter)
		msgAttachment += "Content-Type: text/plain; charset=\"utf-8\"\r\n"
		msgAttachment += "Content-Transfer-Encoding: base64\r\n"
		msgAttachment += "Content-Disposition: attachment;filename=\"" + fmt.Sprintf("%v", data["AttachmentFileName"]) + "\"\r\n"

		//read file
		rawFile, fileErr := ioutil.ReadFile(fmt.Sprintf("%v", data["AttachmentFilePath"]))
		if fileErr != nil {
			log.Panic(fileErr)
		}

		msgAttachment += "\r\n" + base64.StdEncoding.EncodeToString(rawFile)
		body.Write([]byte(msgAttachment))
	}

	err = smtp.SendMail(os.Getenv("SMTP_SERVER"), nil, sender, strings.Split(toList, "; "), body.Bytes())
	if err != nil {
		return "", err
	}

	return "Mail sent.", nil
}
