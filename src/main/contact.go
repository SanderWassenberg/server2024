package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
	SG "github.com/sendgrid/sendgrid-go"
	SGmail "github.com/sendgrid/sendgrid-go/helpers/mail"
)



// Types

type ContactData struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	Email   string `json:"email"`
}



// Vars

// See:
// https://stackoverflow.com/questions/201323/how-can-i-validate-an-email-address-using-a-regular-expression
// https://regexper.com/#(%3F%3A%5Ba-z0-9!%23%24%25%26%27*%2B%2F%3D%3F%5E_%60%7B%7C%7D%7E-%5D%2B(%3F%3A%5C.%5Ba-z0-9!%23%24%25%26%27*%2B%2F%3D%3F%5E_%60%7B%7C%7D%7E-%5D%2B)*%7C%22(%3F%3A%5B%5Cx01-%5Cx08%5Cx0b%5Cx0c%5Cx0e-%5Cx1f%5Cx21%5Cx23-%5Cx5b%5Cx5d-%5Cx7f%5D%7C%5C%5C%5B%5Cx01-%5Cx09%5Cx0b%5Cx0c%5Cx0e-%5Cx7f%5D)*%22)%40(%3F%3A(%3F%3A%5Ba-z0-9%5D(%3F%3A%5Ba-z0-9-%5D*%5Ba-z0-9%5D)%3F%5C.)%2B%5Ba-z0-9%5D(%3F%3A%5Ba-z0-9-%5D*%5Ba-z0-9%5D)%3F%7C%5C%5B(%3F%3A(%3F%3A(2(5%5B0-5%5D%7C%5B0-4%5D%5B0-9%5D)%7C1%5B0-9%5D%5B0-9%5D%7C%5B1-9%5D%3F%5B0-9%5D))%5C.)%7B3%7D(%3F%3A(2(5%5B0-5%5D%7C%5B0-4%5D%5B0-9%5D)%7C1%5B0-9%5D%5B0-9%5D%7C%5B1-9%5D%3F%5B0-9%5D)%7C%5Ba-z0-9-%5D*%5Ba-z0-9%5D%3A(%3F%3A%5B%5Cx01-%5Cx08%5Cx0b%5Cx0c%5Cx0e-%5Cx1f%5Cx21-%5Cx5a%5Cx53-%5Cx7f%5D%7C%5C%5C%5B%5Cx01-%5Cx09%5Cx0b%5Cx0c%5Cx0e-%5Cx7f%5D)%2B)%5C%5D)
// To ensure the whole string is matched and not a portion: https://stackoverflow.com/questions/447250/matching-exact-string-with-javascript
var email_regexp *regexp.Regexp = regexp.MustCompile(`^(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`);



// Utility functions

// returns utf8 length or "way too big" if the byte length is already more than 4x the limit
func verify_utf8_string_len(str string, lim int) (length any, ok bool) {
	if len(str) > lim * 4 { // utf8 uses at most 4 bytes per char, so more bytes than 4*limit is guaranteed to have more characters.
		return "way too big", false
	}
	l := utf8.RuneCountInString(str)
	return l, (l <= lim)
}

func verify_api_key() {
	// How to verify: https://stackoverflow.com/questions/61658558/how-to-test-sendgrid-api-key-is-valid-or-not-without-sending-emails

	log.Println("Verifying api key with Sendrgid servers...")

	req, err := http.NewRequest("GET", "https://api.sendgrid.com/v3/scopes", nil/*body io.Reader*/)
	if err != nil {
		// req.Header.Add("on-behalf-of", "The subuser's username. This header generates the API call as if the subuser account was making the call")
		log.Printf("failed NewRequest(): %v\nValidity of Sendgrid API key remains unknown", err)
		return
	}
	req.Header.Set("authorization", "Bearer " + config.SendGrid_api_key)

	res, err := http.DefaultClient.Do(req);
	if err != nil {
		log.Printf("failed client.Do(): %v\nValidity of Sendgrid API key remains unknown", err)
		return
	}

	defer res.Body.Close()

	api_key_ok := res.StatusCode >= 200 && res.StatusCode < 300

	if api_key_ok {
		log.Print("API key works, mail sending functionality should work.")
	} else {
		log.Printf("WARNING: API key doesn't work, mail sending functionality will not work. (Sedgrid responded with %v %v)", res.StatusCode, http.StatusText(res.StatusCode))
	}
}



// Handlers

func contact_handler(rw http.ResponseWriter, req *http.Request) {

	var cd ContactData
	if err := json.NewDecoder(req.Body).Decode(&cd); err != nil {
		log.Println("Error:", err)
		respond(rw, http.StatusBadRequest, "json decoder error")
		return
	}

	if cd.Subject == "asd" {
		respond(rw, http.StatusOK, "Let's pretend we mailed that one. ;)");
		time.Sleep(2 * time.Second)
		return
	}
	if cd.Subject == "teapot" {
		respond(rw, http.StatusTeapot, "I'm a teapot. C(^)/");
		time.Sleep(2 * time.Second)
		return
	}

	if count, ok := verify_utf8_string_len(cd.Subject, 200); !ok {
		respond_fmt(rw, http.StatusBadRequest, "Subject length should be <200, was %v", count)
		return
	}
	if count, ok := verify_utf8_string_len(cd.Message, 600); !ok {
		respond_fmt(rw, http.StatusBadRequest, "Message length should be <600, was %v", count)
		return
	}
	if count, ok := verify_utf8_string_len(cd.Email, 100); !ok {
		respond_fmt(rw, http.StatusBadRequest, "Email length should be <100, was %v", count)
		return
	}
	if ok := email_regexp.MatchString(cd.Email); !ok {
		respond(rw, http.StatusBadRequest, "Invalid email")
		return
	}

	if ok := send_contact_email(&cd); ok {
		respond(rw, http.StatusOK, "We mailed that bad boi real good!")
	} else {
		respond(rw, http.StatusInternalServerError, "An error occured when sending the email.")
	}
}

func send_contact_email(cd *ContactData) (ok bool) {
	from := SGmail.NewEmail("Showcase website", "sander.wassenberg@windesheim.nl")
	to   := SGmail.NewEmail("Pietje Puk",       "sander.wassenberg@windesheim.nl")

	htmlContent := fmt.Sprintf(
`Hello, <b>Pietje Puk<b>!
You have received a message from somebody on the Showcase website!
<pre>%[1]v</pre>
You can reply to them by sending an email to <a href="mailto:%[2]v">%[2]v</a>
`, html.EscapeString(cd.Message), cd.Email)

	message := SGmail.NewSingleEmail(from, cd.Subject, to, "", htmlContent)
	client  := SG.NewSendClient(config.SendGrid_api_key)
	response, err := client.Send(message)

	if err != nil {
		log.Printf("Error sending email with client.Send(): %v", err)
		return false
	}

	b := strings.Builder{}
	fmt.Fprintf(&b, "Succesfully sent an email!\nResponse: %v %v\nBody:\n`%v`\nHeaders:\n", response.StatusCode, http.StatusText(response.StatusCode), response.Body)
	for k, v := range response.Headers {
		fmt.Fprintf(&b, "- %v: %v\n", k, v)
	}
	log.Print(b.String())

	return true
}