package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"time"
	SG "src/sendgrid/sendgrid-go"
	SGmail "src/sendgrid/sendgrid-go/helpers/mail"
)

// See:
// https://stackoverflow.com/questions/201323/how-can-i-validate-an-email-address-using-a-regular-expression
// https://regexper.com/#(%3F%3A%5Ba-z0-9!%23%24%25%26%27*%2B%2F%3D%3F%5E_%60%7B%7C%7D%7E-%5D%2B(%3F%3A%5C.%5Ba-z0-9!%23%24%25%26%27*%2B%2F%3D%3F%5E_%60%7B%7C%7D%7E-%5D%2B)*%7C%22(%3F%3A%5B%5Cx01-%5Cx08%5Cx0b%5Cx0c%5Cx0e-%5Cx1f%5Cx21%5Cx23-%5Cx5b%5Cx5d-%5Cx7f%5D%7C%5C%5C%5B%5Cx01-%5Cx09%5Cx0b%5Cx0c%5Cx0e-%5Cx7f%5D)*%22)%40(%3F%3A(%3F%3A%5Ba-z0-9%5D(%3F%3A%5Ba-z0-9-%5D*%5Ba-z0-9%5D)%3F%5C.)%2B%5Ba-z0-9%5D(%3F%3A%5Ba-z0-9-%5D*%5Ba-z0-9%5D)%3F%7C%5C%5B(%3F%3A(%3F%3A(2(5%5B0-5%5D%7C%5B0-4%5D%5B0-9%5D)%7C1%5B0-9%5D%5B0-9%5D%7C%5B1-9%5D%3F%5B0-9%5D))%5C.)%7B3%7D(%3F%3A(2(5%5B0-5%5D%7C%5B0-4%5D%5B0-9%5D)%7C1%5B0-9%5D%5B0-9%5D%7C%5B1-9%5D%3F%5B0-9%5D)%7C%5Ba-z0-9-%5D*%5Ba-z0-9%5D%3A(%3F%3A%5B%5Cx01-%5Cx08%5Cx0b%5Cx0c%5Cx0e-%5Cx1f%5Cx21-%5Cx5a%5Cx53-%5Cx7f%5D%7C%5C%5C%5B%5Cx01-%5Cx09%5Cx0b%5Cx0c%5Cx0e-%5Cx7f%5D)%2B)%5C%5D)
// To ensure the whole string is matched and not a portion: https://stackoverflow.com/questions/447250/matching-exact-string-with-javascript
var email_regexp *regexp.Regexp = regexp.MustCompile(`^(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`);

type Contact_Data struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	Email   string `json:"email"`
}


func contact_func(rw http.ResponseWriter, req *http.Request) {
	body, _ := io.ReadAll(req.Body)

	var contact_data Contact_Data
	if err := json.Unmarshal(body, &contact_data); err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		respond(rw, http.StatusBadRequest, err.Error());
		return;
	}

	if contact_data.Subject == "asd" {
		respond(rw, http.StatusOK, "Let's pretend we mailed that one. ;)");
		time.Sleep(2 * time.Second)
		return
	}
	if contact_data.Subject == "teapot" {
		respond(rw, http.StatusTeapot, "Would you like milk with that?");
		time.Sleep(2 * time.Second)
		return
	}

	if msg, args := send_email(&contact_data); msg != "" {
		respond_fmt(rw, http.StatusBadRequest, msg, args...)
		return
	}

	respond(rw, http.StatusOK, "We mailed that bad boi real good!");
	_ = time.Second
}

// return values are effectively an error, but I didn't feel like wrapping them.
func send_email(cd *Contact_Data) (msg string, args []any) {
	if count, ok := verify_utf8_string_len(cd.Subject, 200); !ok {
		return "subject length should be <200, was %v", []any{count}
	}
	if count, ok := verify_utf8_string_len(cd.Message, 600); !ok {
		return "message length should be <600, was %v", []any{count}
	}
	if count, ok := verify_utf8_string_len(cd.Email, 100); !ok {
		return "email length should be <100, was %v", []any{count}
	}
	if ok := email_regexp.MatchString(cd.Email); !ok {
		return "invalid email", nil
	}

	from := SGmail.NewEmail("Showcase website", "sander.wassenberg@windesheim.nl")
	to   := SGmail.NewEmail("Pietje Puk",       "sander.wassenberg@windesheim.nl")

	htmlContent := fmt.Sprint(
`Hello, <b>Pietje Puk<b>!
You have received a message from somebody on the Showcase website!
<pre>%[1]v</pre>
You can reply to them by sending an email to <a href="mailto:%[2]v">%[2]v</a>
`, html.EscapeString(cd.Message), cd.Email)

	message := SGmail.NewSingleEmail(from, cd.Subject, to, "", htmlContent)

	client := SG.NewSendClient(config.SendGrid_api_key)

	response, err := client.Send(message)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Succesfully sent an email!")
		fmt.Println("Response:", response.StatusCode, http.StatusText(response.StatusCode))
		fmt.Println("Body:", response.Body)
		fmt.Println("Headers:", response.Headers)
		for k, v := range response.Headers {
			fmt.Printf("%v: %v\n", k, v)
		}
	}

	return "", nil
}