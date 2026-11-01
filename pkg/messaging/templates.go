package messaging

import "strings"

// Placeholders a message body may use; they are filled in per recipient.
const (
	placeholderName     = "{name}"
	placeholderProperty = "{property}"
	placeholderUnit     = "{unit}"
)

type template struct {
	Body    string `json:"body"`
	Key     string `json:"key"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
}

// templates are the built-in message templates, in the order the UI lists them.
var templates = []template{
	{
		Key:     "general_notice",
		Name:    "General notice",
		Subject: "Notice from your property manager",
		Body:    "Dear {name}, this is a notice regarding {unit} at {property}.",
	},
	{
		Key:     "rent_increase",
		Name:    "Rent increase",
		Subject: "Upcoming change to your rent",
		Body:    "Dear {name}, please note that the rent for {unit} at {property} will be reviewed. Details of the new amount and the date it takes effect will follow.",
	},
	{
		Key:     "scheduled_maintenance",
		Name:    "Scheduled maintenance",
		Subject: "Scheduled maintenance at {property}",
		Body:    "Dear {name}, maintenance work is scheduled at {property}. Access to some facilities near {unit} may be limited during this time.",
	},
	{
		Key:     "payment_reminder",
		Name:    "Payment reminder",
		Subject: "Rent payment reminder",
		Body:    "Dear {name}, this is a reminder that a rent payment for {unit} at {property} is due. Please make your payment to avoid falling behind.",
	},
}

func findTemplate(key string) (template, bool) {
	for _, t := range templates {
		if t.Key == key {
			return t, true
		}
	}
	return template{}, false
}

// personalise fills in a recipient's details. Unknown {tokens} are left as written.
func personalise(text string, name string, unit string, property string) string {
	return strings.NewReplacer(
		placeholderName, name,
		placeholderUnit, unit,
		placeholderProperty, property,
	).Replace(text)
}
