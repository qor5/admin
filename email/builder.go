package email

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/go-faker/faker/v4"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Builder struct {
	db     *gorm.DB
	sender *DefaultSESDriver
}

func ConfigEmailBuilder(db *gorm.DB) *Builder {
	b := New(db).AutoMigrate()
	b.Sender(LoadSenderConfig())
	return b
}

func New(db *gorm.DB) *Builder {
	return &Builder{
		db: db,
	}
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/create":
		if r.Method == http.MethodPost {
			b.createTemplate(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "/list":
		if r.Method == http.MethodGet {
			b.getTemplate(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "/send":
		if r.Method == http.MethodPost {
			b.send(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "page not found", http.StatusNotFound)
	}
}

func (b *Builder) createTemplate(w http.ResponseWriter, r *http.Request) {
	var et MailTemplate
	if err := json.NewDecoder(r.Body).Decode(&et); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := b.db.Create(&et).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	by, _ := json.Marshal(&UnifyResponse{Data: &et})
	_, _ = w.Write(by)
}

func (b *Builder) getTemplate(w http.ResponseWriter, r *http.Request) {
	var templates []MailTemplate
	idsParam := r.URL.Query().Get("ids")
	if strings.TrimSpace(idsParam) == "" {
		if err := b.db.Order("created_at DESC").Limit(10).Find(&templates).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		idsStr := strings.Split(idsParam, ",")
		ids := make([]int, len(idsStr))
		for i, idStr := range idsStr {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "invalid id params", http.StatusBadRequest)
				return
			}
			ids[i] = id
		}

		if err := b.db.Where("id in ?", ids).Find(&templates).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	by, _ := json.Marshal(&UnifyResponse{Data: &templates})
	_, _ = w.Write(by)
}

func (b *Builder) send(w http.ResponseWriter, r *http.Request) {
	var sendRequest SendRequest
	err := json.NewDecoder(r.Body).Decode(&sendRequest)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	var et MailTemplate
	err = b.db.Where("id = ?", sendRequest.TemplateID).First(&et).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "template not found", http.StatusNotFound)

		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	var results []SendResult
	for _, uid := range sendRequest.UserIds {
		//  fake username here(actually, should get it by uid)
		mailData := MailData{
			Name: faker.Name(),
		}
		subjectTmpl, err := GetTemplate(et.Subject)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		subject, err := GetContent(subjectTmpl, mailData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		htmlBodyTmpl, err := GetTemplate(et.HTMLBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		htmlBody, err := GetContent(htmlBodyTmpl, mailData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// get ToEmailAddress by config, actually should get it by uid
		toEmailAddress := LoadToEmailAddress()
		input := &sesv2.SendEmailInput{
			Content: &types.EmailContent{
				Raw: nil,
				Simple: &types.Message{
					Body: &types.Body{
						Html: &types.Content{
							Data:    aws.String(htmlBody),
							Charset: aws.String(b.sender.Config.HTMLBodyCharset),
						},
						Text: nil,
					},
					Subject: &types.Content{
						Data:    aws.String(subject),
						Charset: aws.String(b.sender.Config.SubjectCharset),
					},
					Headers: nil,
				},
				Template: nil,
			},
			ConfigurationSetName: nil,
			Destination: &types.Destination{
				BccAddresses: nil,
				CcAddresses:  nil,
				ToAddresses:  []string{toEmailAddress},
			},
			EmailTags:                      nil,
			EndpointId:                     nil,
			FeedbackForwardingEmailAddress: nil,
			FeedbackForwardingEmailAddressIdentityArn: nil,
			FromEmailAddress: aws.String((&mail.Address{
				Name:    b.sender.Config.FromName,
				Address: b.sender.Config.FromEmailAddress,
			}).String()),
			FromEmailAddressIdentityArn: nil,
			ListManagementOptions:       nil,
			ReplyToAddresses:            nil,
		}
		output, err := b.sender.SendEmail(r.Context(), input)
		if err != nil {
			results = append(results, SendResult{
				UserId:     uid,
				TemplateID: sendRequest.TemplateID,
				MessageID:  "",
				ErrMsg:     err.Error(),
			})
		} else {
			results = append(results, SendResult{
				UserId:     uid,
				TemplateID: sendRequest.TemplateID,
				MessageID:  lo.FromPtr(output.MessageId),
				ErrMsg:     "",
			})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	by, _ := json.Marshal(&UnifyResponse{Data: results})
	_, _ = w.Write(by)
}

func (b *Builder) AutoMigrate() *Builder {
	err := b.db.AutoMigrate(&MailTemplate{})
	if err != nil {
		panic(err)
	}
	return b
}

func (b *Builder) Sender(config SESDriverConfig) *Builder {
	b.sender = NewDefaultSESDriver(config)
	return b
}

type SendRequest struct {
	TemplateID int   `json:"template_id"`
	UserIds    []int `json:"user_ids"`
}

type SendResult struct {
	UserId     int    `json:"user_id"`
	TemplateID int    `json:"template_id"`
	MessageID  string `json:"message_id"`
	ErrMsg     string `json:"err_msg"`
}

type UnifyResponse struct {
	Data interface{} `json:"data"`
}

func LoadToEmailAddress() string {
	to := os.Getenv("TO_ADDRESS")
	if to == "" {
		panic("please set TO_ADDRESS env")
	}
	return to
}

func LoadSenderConfig() (config SESDriverConfig) {
	from := os.Getenv("FROM_ADDRESS")
	if from == "" {
		panic("please set FROM_ADDRESS env")
	}
	return SESDriverConfig{
		FromEmailAddress:               from,
		FromName:                       "ciam",
		SubjectCharset:                 "UTF-8",
		HTMLBodyCharset:                "UTF-8",
		TextBodyCharset:                "UTF-8",
		ConfigurationSetName:           "",
		FeedbackForwardingEmailAddress: "",
		FeedbackForwardingEmailAddressIdentityArn: "",
		FromEmailAddressIdentityArn:               "",
		ContactListName:                           "",
		TopicName:                                 "",
	}
}

type MailData struct {
	Name string
}

func GetTemplate(tmplStr string) (*template.Template, error) {
	tmpl, err := template.New("example").Parse(tmplStr)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func GetContent(tmpl *template.Template, mailData MailData) (string, error) {
	b := bytes.Buffer{}
	err := tmpl.Execute(&b, mailData)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
