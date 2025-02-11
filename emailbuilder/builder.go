package emailbuilder

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/smithy-go"
	"github.com/go-faker/faker/v4"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

var reEmail = regexp.MustCompile(`^[^@]+@[^@]+\.[^@.]+$`)

var EmailValidator = func(email string) bool {
	return reEmail.MatchString(email)
}

type Builder struct {
	db     *gorm.DB
	sender *DefaultSESDriver
}

func ConfigEmailBuilder(db *gorm.DB) *Builder {
	b := New(db).AutoMigrate()
	err := b.PresetData()
	if err != nil {
		log.Println("preset data err:", err)
	} else {
		log.Println("preset data success")
	}
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
	// validate template format
	if _, err := GetTemplate(et.Subject); err != nil {
		http.Error(w, "invalid subject", http.StatusBadRequest)
		return
	}
	// if _, err := GetTemplate(et.JSONBody); err != nil {
	// 	http.Error(w, "invalid json body", http.StatusBadRequest)
	// 	return
	// }
	// if _, err := GetTemplate(et.HTMLBody); err != nil {
	// 	http.Error(w, "invalid html body", http.StatusBadRequest)
	// 	return
	// }
	// save in db
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
	var hasErr bool
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
		toEmailAddress := strings.ToLower(strings.TrimSpace(sendRequest.ToEmailAddress))
		if !EmailValidator(toEmailAddress) {
			http.Error(w, "invalid email address", http.StatusBadRequest)
			return
		}

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
			hasErr = true
			errMsg := "unknown error"
			var opErr *smithy.OperationError
			if errors.As(err, &opErr) {
				var apiErr smithy.APIError
				if errors.As(opErr.Err, &apiErr) {
					errMsg = apiErr.ErrorCode()
				}
			}
			results = append(results, SendResult{
				UserId:     uid,
				TemplateID: sendRequest.TemplateID,
				MessageID:  "",
				ErrMsg:     errMsg,
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
	if hasErr {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
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
	TemplateID     int    `json:"template_id"`
	UserIds        []int  `json:"user_ids"`
	ToEmailAddress string `json:"to_email_address"`
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

func (b *Builder) PresetData() error {
	upsertSql := `INSERT INTO public.mail_templates (id, created_at, updated_at, deleted_at, subject, json_body, html_body)
VALUES (1, '2024-12-26 15:01:13.985362 +00:00', '2024-12-26 15:01:13.985362 +00:00', null, 'Sphero - Newsletter', e'{
  "subject": "Sphero - Newsletter",
  "subTitle": "Nice to meet you!",
  "content": {
    "type": "page",
    "data": {
      "value": {
        "breakpoint": "480px",
        "headAttributes": "",
        "font-size": "14px",
        "font-weight": "400",
        "line-height": "1.7",
        "headStyles": [
          {
            "content": ".mjml-body { width: 600px; margin: 0px auto; }"
          },
          {
            "content": "a {color: inherit} a:hover {color: inherit} a:active {color: inherit}"
          }
        ],
        "fonts": [],
        "responsive": true,
        "font-family": "-apple-system, BlinkMacSystemFont, \'Segoe UI\', \'Roboto\', \'Oxygen\', \'Ubuntu\', \'Cantarell\', \'Fira Sans\', \'Droid Sans\',\'Helvetica Neue\', sans-serif",
        "text-color": "#000000"
      }
    },
    "attributes": {
      "background-color": "#8C9A80",
      "width": "600px",
      "css-class": "mjbody"
    },
    "children": [
      {
        "type": "section",
        "data": {
          "value": {
            "noWrap": false
          }
        },
        "attributes": {
          "padding": "0px 0px 0px 0px",
          "border": "none",
          "direction": "ltr",
          "text-align": "center",
          "background-repeat": "repeat",
          "background-size": "auto",
          "background-position": "top center",
          "background-color": "#FFFFFF",
          "font-family": "Arial, sans-serif"
        },
        "children": [
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "navbar",
                "data": {
                  "value": {
                    "links": [
                      {
                        "color": "#1890ff",
                        "font-size": "13px",
                        "target": "_blank",
                        "font-family": "Arial, sans-serif",
                        "content": "Shop",
                        "padding": "15px 10px 15px 10px"
                      },
                      {
                        "color": "#1890ff",
                        "font-size": "13px",
                        "target": "_blank",
                        "font-family": "Arial, sans-serif",
                        "content": "About",
                        "padding": "15px 10px 15px 10px"
                      },
                      {
                        "color": "#1890ff",
                        "font-size": "13px",
                        "target": "_blank",
                        "font-family": "Arial, sans-serif",
                        "content": "Contact",
                        "padding": "15px 10px 15px 10px"
                      },
                      {
                        "color": "#1890ff",
                        "font-size": "13px",
                        "target": "_blank",
                        "font-family": "Arial, sans-serif",
                        "content": "Blog",
                        "padding": "15px 10px 15px 10px"
                      }
                    ]
                  }
                },
                "attributes": {
                  "align": "center",
                  "font-family": "Arial, sans-serif"
                },
                "children": []
              }
            ]
          }
        ]
      },
      {
        "type": "section",
        "data": {
          "value": {
            "noWrap": false
          }
        },
        "attributes": {
          "background-repeat": "repeat",
          "background-size": "auto",
          "background-position": "top center",
          "border": "none",
          "direction": "ltr",
          "text-align": "center",
          "background-color": "#263D29",
          "font-family": "Arial, sans-serif",
          "padding": "0px 0px 0px 0px"
        },
        "children": [
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "width": "100%",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "text",
                "data": {
                  "value": {
                    "content": "JOIN US FOR A FEAST ON"
                  }
                },
                "attributes": {
                  "align": "left",
                  "color": "#FFFFFF",
                  "font-family": "Arial, sans-serif",
                  "padding": "10px 25px 10px 25px"
                },
                "children": []
              },
              {
                "type": "text",
                "data": {
                  "value": {
                    "content": "St. Patrick\'s{{Name}} Day"
                  }
                },
                "attributes": {
                  "align": "left",
                  "color": "#FFFFFF",
                  "font-family": "Arial, sans - serif",
                  "font-size": "36px",
                  "padding": "10px 25px 10px 25px"
                },
                "children": []
              },
              {
                "type": "image",
                "data": {
                  "value": {}
                },
                "attributes": {
                  "align": "center",
                  "height": "auto",
                  "src": "https://d3k81ch9hvuctc.cloudfront.net/company/S7EvMw/images/29140463-2f83-49ea-922d-ac65063642cd.gif",
                  "font-family": "Arial, sans - serif",
                  "color": "#ffffff",
                  "background-color": "#1A1F25",
                  "grid-color": "#12304b",
                  "shadow-color": "#000000",
                  "padding": "20px 0px 20px 0px"
                },
                "children": [
                  {
                    "type": "text",
                    "data": {
                      "value": {
                        "content": "SUMMER SALE"
                      }
                    },
                    "attributes": {
                      "align": "center",
                      "color": "#ffffff",
                      "font-family": "Arial, sans - serif",
                      "font-size": "32px",
                      "padding": "10px 25px 10px 25px"
                    },
                    "children": []
                  },
                  {
                    "type": "text",
                    "data": {
                      "value": {
                        "content": "<span><em>50% OFF</em></span>"
                      }
                    },
                    "attributes": {
                      "align": "center",
                      "color": "#ffffff",
                      "font-family": "Arial, sans - serif",
                      "font-size": "64px",
                      "padding": "10px 25px 10px 25px"
                    },
                    "children": []
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "section",
        "data": {
          "value": {
            "noWrap": false
          }
        },
        "attributes": {
          "background-repeat": "repeat",
          "background-size": "auto",
          "background-position": "top center",
          "border": "none",
          "direction": "ltr",
          "text-align": "center",
          "background-color": "#263D29",
          "font-family": "Arial, sans - serif",
          "padding": "0px 0px 0px 0px"
        },
        "children": [
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans - serif",
              "width": "480px",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "text",
                "data": {
                  "value": {
                    "content": "From hearty stews and comforting shepherd\'s pie to classic fish and chips, we have something to satisfy every appetite."
                  }
                },
                "attributes": {
                  "align": "left",
                  "color": "#FFFFFF",
                  "font-family": "Arial, sans-serif",
                  "padding": "10px 20px 10px 20px"
                },
                "children": []
              }
            ]
          }
        ]
      },
      {
        "type": "section",
        "data": {
          "value": {
            "noWrap": false
          }
        },
        "attributes": {
          "background-repeat": "repeat",
          "background-size": "auto",
          "background-position": "top center",
          "border": "none",
          "direction": "ltr",
          "text-align": "center",
          "background-color": "#263D29",
          "font-family": "Arial, sans-serif",
          "padding": "0px 0px 0px 0px"
        },
        "children": [
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "width": "100%",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "image",
                "data": {
                  "value": {}
                },
                "attributes": {
                  "align": "center",
                  "height": "auto",
                  "src": "https://d3k81ch9hvuctc.cloudfront.net/company/S7EvMw/images/6644858a-f5af-4bff-917d-f13c5df516ce.png",
                  "font-family": "Arial, sans-serif",
                  "padding": "0px 0px 0px 0px"
                },
                "children": []
              }
            ]
          }
        ]
      },
      {
        "type": "section",
        "data": {
          "value": {
            "noWrap": false
          }
        },
        "attributes": {
          "background-repeat": "repeat",
          "background-size": "auto",
          "background-position": "top center",
          "border": "none",
          "direction": "ltr",
          "text-align": "center",
          "background-color": "#263D29",
          "font-family": "Arial, sans-serif",
          "padding": "0px 0px 0px 0px"
        },
        "children": [
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "width": "100%",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "button",
                "data": {
                  "value": {
                    "content": "Book a table"
                  }
                },
                "attributes": {
                  "align": "center",
                  "background-color": "#C5900C",
                  "color": "#ffffff",
                  "font-weight": "normal",
                  "border-radius": "3px",
                  "line-height": "120%",
                  "target": "_blank",
                  "vertical-align": "middle",
                  "border": "none",
                  "text-align": "center",
                  "href": "#",
                  "font-family": "Arial, sans-serif",
                  "font-size": "16px",
                  "width": "100%",
                  "padding": "10px 25px 10px 25px"
                },
                "children": []
              },
              {
                "type": "button",
                "data": {
                  "value": {
                    "content": "Book a table"
                  }
                },
                "attributes": {
                  "align": "center",
                  "background-color": "#FCF0D3",
                  "color": "#C5900C",
                  "font-weight": "normal",
                  "border-radius": "3px",
                  "line-height": "120%",
                  "target": "_blank",
                  "vertical-align": "middle",
                  "border": "3px solid #C5900C",
                  "text-align": "center",
                  "href": "#",
                  "font-family": "Arial, sans-serif",
                  "font-size": "16px",
                  "width": "100%",
                  "padding": "10px 25px 10px 25px"
                },
                "children": []
              }
            ]
          }
        ]
      },
      {
        "type": "section",
        "data": {
          "value": {
            "noWrap": false
          }
        },
        "attributes": {
          "background-repeat": "repeat",
          "background-size": "auto",
          "background-position": "top center",
          "border": "none",
          "direction": "ltr",
          "text-align": "center",
          "background-color": "#FFFFFF",
          "font-family": "Arial, sans-serif",
          "padding": "0px 0px 0px 0px"
        },
        "children": [
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "width": "50%",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "text",
                "data": {
                  "value": {
                    "content": "And of course, no Irish meal would be complete without a pint of Guinness or a dram of Irish whiskey to wash it down.<br /><br /><span><strong>View drink menu</strong></span>"
                  }
                },
                "attributes": {
                  "align": "left",
                  "font-family": "Arial, sans-serif",
                  "padding": "25px 25px 25px 25px"
                },
                "children": []
              }
            ]
          },
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "width": "50%",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "image",
                "data": {
                  "value": {}
                },
                "attributes": {
                  "align": "center",
                  "height": "auto",
                  "src": "https://d3k81ch9hvuctc.cloudfront.net/company/S7EvMw/images/a096e725-a517-477e-bdb9-97211537aeb8.png",
                  "font-family": "Arial, sans-serif",
                  "padding": "25px 25px 25px 25px"
                },
                "children": []
              }
            ]
          }
        ]
      },
      {
        "type": "section",
        "data": {
          "value": {
            "noWrap": false
          }
        },
        "attributes": {
          "background-repeat": "repeat",
          "background-size": "auto",
          "background-position": "top center",
          "border": "none",
          "direction": "ltr",
          "text-align": "center",
          "background-color": "#ffffff",
          "font-family": "Arial, sans-serif",
          "padding": "0px 0px 0px 0px"
        },
        "children": [
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "width": "100%",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "divider",
                "data": {
                  "value": {}
                },
                "attributes": {
                  "align": "center",
                  "border-width": "1px",
                  "border-style": "solid",
                  "border-color": "#C9CCCF",
                  "font-family": "Arial, sans-serif",
                  "padding": "10px 0px 10px 0px"
                },
                "children": []
              }
            ]
          }
        ]
      },
      {
        "type": "section",
        "data": {
          "value": {
            "noWrap": false
          }
        },
        "attributes": {
          "background-repeat": "repeat",
          "background-size": "auto",
          "background-position": "top center",
          "border": "none",
          "direction": "rtl",
          "text-align": "center",
          "background-color": "#FFFFFF",
          "font-family": "Arial, sans-serif",
          "padding": "0px 0px 0px 0px"
        },
        "children": [
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "width": "50%",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "text",
                "data": {
                  "value": {
                    "content": "And of course, no Irish meal would be complete without a pint of Guinness or a dram of Irish whiskey to wash it down.<br /><br /><span><strong>View drink menu</strong></span>"
                  }
                },
                "attributes": {
                  "align": "left",
                  "font-family": "Arial, sans-serif",
                  "padding": "25px 25px 25px 25px"
                },
                "children": []
              }
            ]
          },
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "width": "50%",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "image",
                "data": {
                  "value": {}
                },
                "attributes": {
                  "align": "center",
                  "height": "auto",
                  "src": "https://d3k81ch9hvuctc.cloudfront.net/company/S7EvMw/images/f812b491-36cb-4a82-b2f0-eea60ed5eba9.png",
                  "font-family": "Arial, sans-serif",
                  "padding": "25px 25px 25px 25px"
                },
                "children": []
              }
            ]
          }
        ]
      },
      {
        "type": "section",
        "data": {
          "value": {
            "noWrap": false
          }
        },
        "attributes": {
          "background-repeat": "repeat",
          "background-size": "auto",
          "background-position": "top center",
          "border": "none",
          "direction": "ltr",
          "text-align": "center",
          "background-color": "#263D29",
          "font-family": "Arial, sans-serif",
          "padding": "0px 0px 0px 0px"
        },
        "children": [
          {
            "type": "column",
            "data": {
              "value": {}
            },
            "attributes": {
              "border": "none",
              "vertical-align": "top",
              "font-family": "Arial, sans-serif",
              "width": "100%",
              "padding": "0px 0px 0px 0px"
            },
            "children": [
              {
                "type": "text",
                "data": {
                  "value": {
                    "content": "No longer want to receive these emails? <span><a href=\\"#\\" target=\\"_blank\\">unsubscribe</a></span>"
                  }
                },
                "attributes": {
                  "align": "left",
                  "color": "#FFFFFF",
                  "font-family": "Arial, sans-serif",
                  "padding": "10px 25px 10px 25px"
                },
                "children": []
              }
            ]
          }
        ]
      }
    ]
  }
}', e'<!doctype html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">
  <head>
    <title>

    </title>
    <!--[if !mso]><!-->
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <!--<![endif]-->
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style type="text/css">
      #outlook a { padding:0; }
      body { margin:0;padding:0;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%; }
      table, td { border-collapse:collapse;mso-table-lspace:0pt;mso-table-rspace:0pt; }
      img { border:0;height:auto;line-height:100%; outline:none;text-decoration:none;-ms-interpolation-mode:bicubic; }
      p { display:block;margin:13px 0; }
    </style>
    <!--[if mso]>
    <noscript>
    <xml>
    <o:OfficeDocumentSettings>
      <o:AllowPNG/>
      <o:PixelsPerInch>96</o:PixelsPerInch>
    </o:OfficeDocumentSettings>
    </xml>
    </noscript>
    <![endif]-->
    <!--[if lte mso 11]>
    <style type="text/css">
      .mj-outlook-group-fix { width:100% !important; }
    </style>
    <![endif]-->


    <style type="text/css">
      @media only screen and (min-width:480px) {
        .mj-column-per-100 { width:100% !important; max-width: 100%; }
.mj-column-px-480 { width:480px !important; max-width: 480px; }
.mj-column-per-50 { width:50% !important; max-width: 50%; }
      }
    </style>
    <style media="screen and (min-width:480px)">
      .moz-text-html .mj-column-per-100 { width:100% !important; max-width: 100%; }
.moz-text-html .mj-column-px-480 { width:480px !important; max-width: 480px; }
.moz-text-html .mj-column-per-50 { width:50% !important; max-width: 50%; }
    </style>


    <style type="text/css">



      noinput.mj-menu-checkbox { display:block!important; max-height:none!important; visibility:visible!important; }

      @media only screen and (max-width:480px) {
        .mj-menu-checkbox[type="checkbox"] ~ .mj-inline-links { display:none!important; }
        .mj-menu-checkbox[type="checkbox"]:checked ~ .mj-inline-links,
        .mj-menu-checkbox[type="checkbox"] ~ .mj-menu-trigger { display:block!important; max-width:none!important; max-height:none!important; font-size:inherit!important; }
        .mj-menu-checkbox[type="checkbox"] ~ .mj-inline-links > a { display:block!important; }
        .mj-menu-checkbox[type="checkbox"]:checked ~ .mj-menu-trigger .mj-menu-icon-close { display:block!important; }
        .mj-menu-checkbox[type="checkbox"]:checked ~ .mj-menu-trigger .mj-menu-icon-open { display:none!important; }
      }


    @media only screen and (max-width:480px) {
      table.mj-full-width-mobile { width: 100% !important; }
      td.mj-full-width-mobile { width: auto !important; }
    }

    </style>
    <style type="text/css">
    .mjml-body { width: 600px; margin: 0px auto; }a {color: inherit} a:hover {color: inherit} a:active {color: inherit}
    </style>

  </head>
  <body style="word-spacing:normal;background-color:#8C9A80;">


      <div
         class="mjbody" style="background-color:#8C9A80;"
      >


      <!--[if mso | IE]><table align="center" border="0" cellpadding="0" cellspacing="0" class="email-block-outlook node-idx-content.children.[0]-outlook node-type-section-outlook" role="presentation" style="width:600px;" width="600" bgcolor="#FFFFFF" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->


      <div  class="email-block node-idx-content.children.[0] node-type-section" style="background:#FFFFFF;background-color:#FFFFFF;margin:0px auto;max-width:600px;">

        <table
           align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#FFFFFF;background-color:#FFFFFF;width:100%;"
        >
          <tbody>
            <tr>
              <td
                 style="border:none;direction:ltr;font-size:0px;padding:0px 0px 0px 0px;text-align:center;"
              >
                <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="email-block-outlook node-idx-content.children.[0].children.[0]-outlook node-type-column-outlook" style="vertical-align:top;width:600px;" ><![endif]-->

      <div
         class="mj-column-per-100 mj-outlook-group-fix email-block node-idx-content.children.[0].children.[0] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="center" class="email-block node-idx-content.children.[0].children.[0].children.[0] node-type-navbar" style="font-size:0px;word-break:break-word;"
                >


        <div
           class="mj-inline-links" style=""
        >

    <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0" align="center"><tr><td style="padding:15px 10px 15px 10px;" class="" ><![endif]-->


      <a
         class="mj-link" target="_blank" style="display:inline-block;color:#1890ff;font-family:Arial, sans-serif;font-size:13px;font-weight:normal;line-height:22px;text-decoration:none;text-transform:uppercase;padding:15px 10px 15px 10px;"
      >
        Shop
      </a>


    <!--[if mso | IE]></td><td style="padding:15px 10px 15px 10px;" class="" ><![endif]-->


      <a
         class="mj-link" target="_blank" style="display:inline-block;color:#1890ff;font-family:Arial, sans-serif;font-size:13px;font-weight:normal;line-height:22px;text-decoration:none;text-transform:uppercase;padding:15px 10px 15px 10px;"
      >
        About
      </a>


    <!--[if mso | IE]></td><td style="padding:15px 10px 15px 10px;" class="" ><![endif]-->


      <a
         class="mj-link" target="_blank" style="display:inline-block;color:#1890ff;font-family:Arial, sans-serif;font-size:13px;font-weight:normal;line-height:22px;text-decoration:none;text-transform:uppercase;padding:15px 10px 15px 10px;"
      >
        Contact
      </a>


    <!--[if mso | IE]></td><td style="padding:15px 10px 15px 10px;" class="" ><![endif]-->


      <a
         class="mj-link" target="_blank" style="display:inline-block;color:#1890ff;font-family:Arial, sans-serif;font-size:13px;font-weight:normal;line-height:22px;text-decoration:none;text-transform:uppercase;padding:15px 10px 15px 10px;"
      >
        Blog
      </a>


    <!--[if mso | IE]></td></tr></table><![endif]-->

        </div>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td></tr></table><![endif]-->
              </td>
            </tr>
          </tbody>
        </table>

      </div>


      <!--[if mso | IE]></td></tr></table><table align="center" border="0" cellpadding="0" cellspacing="0" class="email-block-outlook node-idx-content.children.[1]-outlook node-type-section-outlook" role="presentation" style="width:600px;" width="600" bgcolor="#263D29" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->


      <div  class="email-block node-idx-content.children.[1] node-type-section" style="background:#263D29;background-color:#263D29;margin:0px auto;max-width:600px;">

        <table
           align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#263D29;background-color:#263D29;width:100%;"
        >
          <tbody>
            <tr>
              <td
                 style="border:none;direction:ltr;font-size:0px;padding:0px 0px 0px 0px;text-align:center;"
              >
                <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="email-block-outlook node-idx-content.children.[1].children.[0]-outlook node-type-column-outlook" style="vertical-align:top;width:600px;" ><![endif]-->

      <div
         class="mj-column-per-100 mj-outlook-group-fix email-block node-idx-content.children.[1].children.[0] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="left" class="email-block node-idx-content.children.[1].children.[0].children.[0] node-type-text" style="font-size:0px;padding:10px 25px 10px 25px;word-break:break-word;"
                >

      <div
         style="font-family:Arial, sans-serif;font-size:14px;font-weight:400;line-height:1.7;text-align:left;color:#FFFFFF;"
      >JOIN US FOR A FEAST ON</div>

                </td>
              </tr>

              <tr>
                <td
                   align="left" class="email-block node-idx-content.children.[1].children.[0].children.[1] node-type-text" style="font-size:0px;padding:10px 25px 10px 25px;word-break:break-word;"
                >

      <div
         style="font-family:Arial, sans-serif;font-size:36px;font-weight:400;line-height:1.7;text-align:left;color:#FFFFFF;"
      >St. Patrick\'s{{.Name}} Day</div>

                </td>
              </tr>

              <tr>
                <td
                   align="center" class="email-block node-idx-content.children.[1].children.[0].children.[2] node-type-image" style="font-size:0px;padding:20px 0px 20px 0px;word-break:break-word;"
                >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:collapse;border-spacing:0px;"
      >
        <tbody>
          <tr>
            <td  style="width:600px;">

      <img
         height="auto" src="https://d3k81ch9hvuctc.cloudfront.net/company/S7EvMw/images/29140463-2f83-49ea-922d-ac65063642cd.gif" style="border:0;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;" width="600"
      />

            </td>
          </tr>
        </tbody>
      </table>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td></tr></table><![endif]-->
              </td>
            </tr>
          </tbody>
        </table>

      </div>


      <!--[if mso | IE]></td></tr></table><table align="center" border="0" cellpadding="0" cellspacing="0" class="email-block-outlook node-idx-content.children.[2]-outlook node-type-section-outlook" role="presentation" style="width:600px;" width="600" bgcolor="#263D29" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->


      <div  class="email-block node-idx-content.children.[2] node-type-section" style="background:#263D29;background-color:#263D29;margin:0px auto;max-width:600px;">

        <table
           align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#263D29;background-color:#263D29;width:100%;"
        >
          <tbody>
            <tr>
              <td
                 style="border:none;direction:ltr;font-size:0px;padding:0px 0px 0px 0px;text-align:center;"
              >
                <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="email-block-outlook node-idx-content.children.[2].children.[0]-outlook node-type-column-outlook" style="vertical-align:top;width:480px;" ><![endif]-->

      <div
         class="mj-column-px-480 mj-outlook-group-fix email-block node-idx-content.children.[2].children.[0] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="left" class="email-block node-idx-content.children.[2].children.[0].children.[0] node-type-text" style="font-size:0px;padding:10px 20px 10px 20px;word-break:break-word;"
                >

      <div
         style="font-family:Arial, sans-serif;font-size:14px;font-weight:400;line-height:1.7;text-align:left;color:#FFFFFF;"
      >From hearty stews and comforting shepherd\'s pie to classic fish and chips, we have something to satisfy every appetite.</div>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td></tr></table><![endif]-->
              </td>
            </tr>
          </tbody>
        </table>

      </div>


      <!--[if mso | IE]></td></tr></table><table align="center" border="0" cellpadding="0" cellspacing="0" class="email-block-outlook node-idx-content.children.[3]-outlook node-type-section-outlook" role="presentation" style="width:600px;" width="600" bgcolor="#263D29" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->


      <div  class="email-block node-idx-content.children.[3] node-type-section" style="background:#263D29;background-color:#263D29;margin:0px auto;max-width:600px;">

        <table
           align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#263D29;background-color:#263D29;width:100%;"
        >
          <tbody>
            <tr>
              <td
                 style="border:none;direction:ltr;font-size:0px;padding:0px 0px 0px 0px;text-align:center;"
              >
                <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="email-block-outlook node-idx-content.children.[3].children.[0]-outlook node-type-column-outlook" style="vertical-align:top;width:600px;" ><![endif]-->

      <div
         class="mj-column-per-100 mj-outlook-group-fix email-block node-idx-content.children.[3].children.[0] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="center" class="email-block node-idx-content.children.[3].children.[0].children.[0] node-type-image" style="font-size:0px;padding:0px 0px 0px 0px;word-break:break-word;"
                >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:collapse;border-spacing:0px;"
      >
        <tbody>
          <tr>
            <td  style="width:600px;">

      <img
         height="auto" src="https://d3k81ch9hvuctc.cloudfront.net/company/S7EvMw/images/6644858a-f5af-4bff-917d-f13c5df516ce.png" style="border:0;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;" width="600"
      />

            </td>
          </tr>
        </tbody>
      </table>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td></tr></table><![endif]-->
              </td>
            </tr>
          </tbody>
        </table>

      </div>


      <!--[if mso | IE]></td></tr></table><table align="center" border="0" cellpadding="0" cellspacing="0" class="email-block-outlook node-idx-content.children.[4]-outlook node-type-section-outlook" role="presentation" style="width:600px;" width="600" bgcolor="#263D29" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->


      <div  class="email-block node-idx-content.children.[4] node-type-section" style="background:#263D29;background-color:#263D29;margin:0px auto;max-width:600px;">

        <table
           align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#263D29;background-color:#263D29;width:100%;"
        >
          <tbody>
            <tr>
              <td
                 style="border:none;direction:ltr;font-size:0px;padding:0px 0px 0px 0px;text-align:center;"
              >
                <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="email-block-outlook node-idx-content.children.[4].children.[0]-outlook node-type-column-outlook" style="vertical-align:top;width:600px;" ><![endif]-->

      <div
         class="mj-column-per-100 mj-outlook-group-fix email-block node-idx-content.children.[4].children.[0] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="center" vertical-align="middle" class="email-block node-idx-content.children.[4].children.[0].children.[0] node-type-button" style="font-size:0px;padding:10px 25px 10px 25px;word-break:break-word;"
                >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:separate;width:100%;line-height:100%;"
      >
        <tbody>
          <tr>
            <td
               align="center" bgcolor="#C5900C" role="presentation" style="border:none;border-radius:3px;cursor:auto;mso-padding-alt:10px 25px;text-align:center;background:#C5900C;" valign="middle"
            >
              <a
                 href="#" style="display:inline-block;background:#C5900C;color:#ffffff;font-family:Arial, sans-serif;font-size:16px;font-weight:normal;line-height:120%;margin:0;text-decoration:none;text-transform:none;padding:10px 25px;mso-padding-alt:0px;border-radius:3px;" target="_blank"
              >
                Book a table
              </a>
            </td>
          </tr>
        </tbody>
      </table>

                </td>
              </tr>

              <tr>
                <td
                   align="center" vertical-align="middle" class="email-block node-idx-content.children.[4].children.[0].children.[1] node-type-button" style="font-size:0px;padding:10px 25px 10px 25px;word-break:break-word;"
                >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:separate;width:100%;line-height:100%;"
      >
        <tbody>
          <tr>
            <td
               align="center" bgcolor="#FCF0D3" role="presentation" style="border:3px solid #C5900C;border-radius:3px;cursor:auto;mso-padding-alt:10px 25px;text-align:center;background:#FCF0D3;" valign="middle"
            >
              <a
                 href="#" style="display:inline-block;background:#FCF0D3;color:#C5900C;font-family:Arial, sans-serif;font-size:16px;font-weight:normal;line-height:120%;margin:0;text-decoration:none;text-transform:none;padding:10px 25px;mso-padding-alt:0px;border-radius:3px;" target="_blank"
              >
                Book a table
              </a>
            </td>
          </tr>
        </tbody>
      </table>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td></tr></table><![endif]-->
              </td>
            </tr>
          </tbody>
        </table>

      </div>


      <!--[if mso | IE]></td></tr></table><table align="center" border="0" cellpadding="0" cellspacing="0" class="email-block-outlook node-idx-content.children.[5]-outlook node-type-section-outlook" role="presentation" style="width:600px;" width="600" bgcolor="#FFFFFF" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->


      <div  class="email-block node-idx-content.children.[5] node-type-section" style="background:#FFFFFF;background-color:#FFFFFF;margin:0px auto;max-width:600px;">

        <table
           align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#FFFFFF;background-color:#FFFFFF;width:100%;"
        >
          <tbody>
            <tr>
              <td
                 style="border:none;direction:ltr;font-size:0px;padding:0px 0px 0px 0px;text-align:center;"
              >
                <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="email-block-outlook node-idx-content.children.[5].children.[0]-outlook node-type-column-outlook" style="vertical-align:top;width:300px;" ><![endif]-->

      <div
         class="mj-column-per-50 mj-outlook-group-fix email-block node-idx-content.children.[5].children.[0] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="left" class="email-block node-idx-content.children.[5].children.[0].children.[0] node-type-text" style="font-size:0px;padding:25px 25px 25px 25px;word-break:break-word;"
                >

      <div
         style="font-family:Arial, sans-serif;font-size:14px;font-weight:400;line-height:1.7;text-align:left;color:#000000;"
      >And of course, no Irish meal would be complete without a pint of Guinness or a dram of Irish whiskey to wash it down.<br /><br /><span><strong>View drink menu</strong></span></div>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td><td class="email-block-outlook node-idx-content.children.[5].children.[1]-outlook node-type-column-outlook" style="vertical-align:top;width:300px;" ><![endif]-->

      <div
         class="mj-column-per-50 mj-outlook-group-fix email-block node-idx-content.children.[5].children.[1] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="center" class="email-block node-idx-content.children.[5].children.[1].children.[0] node-type-image" style="font-size:0px;padding:25px 25px 25px 25px;word-break:break-word;"
                >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:collapse;border-spacing:0px;"
      >
        <tbody>
          <tr>
            <td  style="width:250px;">

      <img
         height="auto" src="https://d3k81ch9hvuctc.cloudfront.net/company/S7EvMw/images/a096e725-a517-477e-bdb9-97211537aeb8.png" style="border:0;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;" width="250"
      />

            </td>
          </tr>
        </tbody>
      </table>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td></tr></table><![endif]-->
              </td>
            </tr>
          </tbody>
        </table>

      </div>


      <!--[if mso | IE]></td></tr></table><table align="center" border="0" cellpadding="0" cellspacing="0" class="email-block-outlook node-idx-content.children.[6]-outlook node-type-section-outlook" role="presentation" style="width:600px;" width="600" bgcolor="#ffffff" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->


      <div  class="email-block node-idx-content.children.[6] node-type-section" style="background:#ffffff;background-color:#ffffff;margin:0px auto;max-width:600px;">

        <table
           align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#ffffff;background-color:#ffffff;width:100%;"
        >
          <tbody>
            <tr>
              <td
                 style="border:none;direction:ltr;font-size:0px;padding:0px 0px 0px 0px;text-align:center;"
              >
                <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="email-block-outlook node-idx-content.children.[6].children.[0]-outlook node-type-column-outlook" style="vertical-align:top;width:600px;" ><![endif]-->

      <div
         class="mj-column-per-100 mj-outlook-group-fix email-block node-idx-content.children.[6].children.[0] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="center" class="email-block node-idx-content.children.[6].children.[0].children.[0] node-type-divider" style="font-size:0px;padding:10px 0px 10px 0px;word-break:break-word;"
                >

      <p
         style="border-top:solid 1px #C9CCCF;font-size:1px;margin:0px auto;width:100%;"
      >
      </p>

      <!--[if mso | IE]><table align="center" border="0" cellpadding="0" cellspacing="0" style="border-top:solid 1px #C9CCCF;font-size:1px;margin:0px auto;width:600px;" role="presentation" width="600px" ><tr><td style="height:0;line-height:0;"> &nbsp;
</td></tr></table><![endif]-->


                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td></tr></table><![endif]-->
              </td>
            </tr>
          </tbody>
        </table>

      </div>


      <!--[if mso | IE]></td></tr></table><table align="center" border="0" cellpadding="0" cellspacing="0" class="email-block-outlook node-idx-content.children.[7]-outlook node-type-section-outlook" role="presentation" style="width:600px;" width="600" bgcolor="#FFFFFF" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->


      <div  class="email-block node-idx-content.children.[7] node-type-section" style="background:#FFFFFF;background-color:#FFFFFF;margin:0px auto;max-width:600px;">

        <table
           align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#FFFFFF;background-color:#FFFFFF;width:100%;"
        >
          <tbody>
            <tr>
              <td
                 style="border:none;direction:rtl;font-size:0px;padding:0px 0px 0px 0px;text-align:center;"
              >
                <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="email-block-outlook node-idx-content.children.[7].children.[0]-outlook node-type-column-outlook" style="vertical-align:top;width:300px;" ><![endif]-->

      <div
         class="mj-column-per-50 mj-outlook-group-fix email-block node-idx-content.children.[7].children.[0] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="left" class="email-block node-idx-content.children.[7].children.[0].children.[0] node-type-text" style="font-size:0px;padding:25px 25px 25px 25px;word-break:break-word;"
                >

      <div
         style="font-family:Arial, sans-serif;font-size:14px;font-weight:400;line-height:1.7;text-align:left;color:#000000;"
      >And of course, no Irish meal would be complete without a pint of Guinness or a dram of Irish whiskey to wash it down.<br /><br /><span><strong>View drink menu</strong></span></div>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td><td class="email-block-outlook node-idx-content.children.[7].children.[1]-outlook node-type-column-outlook" style="vertical-align:top;width:300px;" ><![endif]-->

      <div
         class="mj-column-per-50 mj-outlook-group-fix email-block node-idx-content.children.[7].children.[1] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="center" class="email-block node-idx-content.children.[7].children.[1].children.[0] node-type-image" style="font-size:0px;padding:25px 25px 25px 25px;word-break:break-word;"
                >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="border-collapse:collapse;border-spacing:0px;"
      >
        <tbody>
          <tr>
            <td  style="width:250px;">

      <img
         height="auto" src="https://d3k81ch9hvuctc.cloudfront.net/company/S7EvMw/images/f812b491-36cb-4a82-b2f0-eea60ed5eba9.png" style="border:0;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;" width="250"
      />

            </td>
          </tr>
        </tbody>
      </table>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td></tr></table><![endif]-->
              </td>
            </tr>
          </tbody>
        </table>

      </div>


      <!--[if mso | IE]></td></tr></table><table align="center" border="0" cellpadding="0" cellspacing="0" class="email-block-outlook node-idx-content.children.[8]-outlook node-type-section-outlook" role="presentation" style="width:600px;" width="600" bgcolor="#263D29" ><tr><td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;"><![endif]-->


      <div  class="email-block node-idx-content.children.[8] node-type-section" style="background:#263D29;background-color:#263D29;margin:0px auto;max-width:600px;">

        <table
           align="center" border="0" cellpadding="0" cellspacing="0" role="presentation" style="background:#263D29;background-color:#263D29;width:100%;"
        >
          <tbody>
            <tr>
              <td
                 style="border:none;direction:ltr;font-size:0px;padding:0px 0px 0px 0px;text-align:center;"
              >
                <!--[if mso | IE]><table role="presentation" border="0" cellpadding="0" cellspacing="0"><tr><td class="email-block-outlook node-idx-content.children.[8].children.[0]-outlook node-type-column-outlook" style="vertical-align:top;width:600px;" ><![endif]-->

      <div
         class="mj-column-per-100 mj-outlook-group-fix email-block node-idx-content.children.[8].children.[0] node-type-column" style="font-size:0px;text-align:left;direction:ltr;display:inline-block;vertical-align:top;width:100%;"
      >

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%"
      >
        <tbody>
          <tr>
            <td  style="border:none;vertical-align:top;padding:0px 0px 0px 0px;">

      <table
         border="0" cellpadding="0" cellspacing="0" role="presentation" style="" width="100%"
      >
        <tbody>

              <tr>
                <td
                   align="left" class="email-block node-idx-content.children.[8].children.[0].children.[0] node-type-text" style="font-size:0px;padding:10px 25px 10px 25px;word-break:break-word;"
                >

      <div
         style="font-family:Arial, sans-serif;font-size:14px;font-weight:400;line-height:1.7;text-align:left;color:#FFFFFF;"
      >No longer want to receive these emails? <span><a href="#" target="_blank">unsubscribe</a></span></div>

                </td>
              </tr>

        </tbody>
      </table>

            </td>
          </tr>
        </tbody>
      </table>

      </div>

          <!--[if mso | IE]></td></tr></table><![endif]-->
              </td>
            </tr>
          </tbody>
        </table>

      </div>


      <!--[if mso | IE]></td></tr></table><![endif]-->


      </div>

  </body>
</html>
  ') ON CONFLICT (id)
DO
UPDATE SET
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at,
    subject = EXCLUDED.subject,
    json_body = EXCLUDED.json_body,
    html_body = EXCLUDED.html_body;
`

	return b.db.Exec(upsertSql).Error
}
