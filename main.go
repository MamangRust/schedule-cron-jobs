package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"text/template"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// Mock booking structure
type Booking struct {
	OrderID     string
	UserEmail   string
	BookingTime time.Time
	CheckInTime time.Time
}

// Mock booking repository
type BookingRepository struct{}

// Dummy method to simulate fetching bookings for a specific time
func (repo *BookingRepository) FindBookingsByBookingTime(bookingTime time.Time) []Booking {
	// Simulating found bookings
	return []Booking{
		{OrderID: "ORD123", UserEmail: "user1@example.com", BookingTime: bookingTime, CheckInTime: bookingTime.Add(1 * time.Hour)},
		{OrderID: "ORD456", UserEmail: "user2@example.com", BookingTime: bookingTime, CheckInTime: bookingTime.Add(2 * time.Hour)},
	}
}

// Mock booking service
type BookingService struct {
	Repo *BookingRepository
}

// Email configuration (can be customized)
const (
	smtpHost     = "smtp.ethereal.email"
	smtpPort     = "587"
	smtpUsername = "rosamond.nienow@ethereal.email"
	smtpPassword = "8N7SxafeM2ZkhefVfp"
	fromEmail    = "rosamond.nienow@ethereal.email"
)

// Function to send an email
type EmailData struct {
	OrderID     string
	UserEmail   string
	BookingTime string
	CheckInTime string
}

// Updated sendEmail function
func sendEmail(booking Booking) error {
	// SMTP setup
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	// Prepare the email subject
	subject := "Booking Confirmation for Order " + booking.OrderID

	// Prepare the email template
	emailTemplate := `
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";

<html>
<body>
    <h2>Booking Confirmation</h2>
    <p>Dear Customer,</p>
    <p>Your booking has been confirmed with the following details:</p>
    <ul>
        <li>Order ID: {{.OrderID}}</li>
        <li>User Email: {{.UserEmail}}</li>
        <li>Booking Time: {{.BookingTime}}</li>
        <li>Check-in Time: {{.CheckInTime}}</li>
    </ul>
    <p>Thank you for your reservation!</p>
</body>
</html>
`

	// Parse the template
	tmpl, err := template.New("emailTemplate").Parse(emailTemplate)
	if err != nil {
		return fmt.Errorf("error parsing email template: %v", err)
	}

	// Prepare the data to inject into the template
	emailData := EmailData{
		OrderID:     booking.OrderID,
		UserEmail:   booking.UserEmail,
		BookingTime: booking.BookingTime.Format(time.RFC1123),
		CheckInTime: booking.CheckInTime.Format(time.RFC1123),
	}

	// Buffer to hold the parsed HTML
	var body bytes.Buffer

	// Write the email headers
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))

	// Apply the parsed template to the emailData and write the result to the buffer
	err = tmpl.Execute(&body, emailData)
	if err != nil {
		return fmt.Errorf("error executing email template: %v", err)
	}

	// Send the email using the SMTP server
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, []string{booking.UserEmail}, body.Bytes())
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	log.Printf("Email sent successfully to %s for Order ID %s", booking.UserEmail, booking.OrderID)
	return nil
}

// Method to process bookings for a specific time
func (service *BookingService) ProcessBookingsForTime(bookingTime time.Time) error {
	bookings := service.Repo.FindBookingsByBookingTime(bookingTime)

	if len(bookings) == 0 {
		log.Println("No bookings found for: ", bookingTime)
		return nil
	}

	for _, booking := range bookings {
		log.Printf("Processing booking: Order ID %s, User Email: %s", booking.OrderID, booking.UserEmail)

		err := sendEmail(booking)
		if err != nil {
			log.Printf("Failed to send email for Order ID %s: %v", booking.OrderID, err)
		} else {
			log.Printf("Successfully sent email for Order ID %s", booking.OrderID)
		}
	}
	return nil
}

func main() {
	// Initialize mock repository and service
	repo := &BookingRepository{}
	bookingService := &BookingService{Repo: repo}

	// Initialize scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal("Error initializing scheduler: ", err)
	}

	defer func() {
		_ = s.Shutdown()
	}()

	_, err = s.NewJob(
		gocron.CronJob(
			"0 58 19 * * *",
			true, // Include seconds
		),
		gocron.NewTask(
			func() {
				currentTime := time.Now()
				log.Printf("Starting to process bookings for: %s", currentTime)

				err := bookingService.ProcessBookingsForTime(currentTime)
				if err != nil {
					log.Println("Error processing bookings:", err)
				} else {
					log.Println("Successfully processed bookings for:", currentTime)
				}

				log.Printf("Finished processing bookings for: %s", currentTime)
			},
		),
	)

	if err != nil {
		log.Fatal("Error creating job: ", err)
	}

	log.Println("Booking service started. Press Ctrl+C to exit.")
	log.Println("Scheduler will run daily at 7:58 PM.")

	// Start the scheduler
	s.Start()

	// Keep the program running
	select {}
}
