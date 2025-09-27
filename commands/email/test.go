package email

import (
	"bctbackend/commands/common"
	"net/smtp"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type TestEmailCommand struct {
	common.Command
}

func NewTestEmailCommand() *cobra.Command {
	var command *TestEmailCommand

	command = &TestEmailCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "test",
				Short: "Send a test email to the administrator",
				Long: heredoc.Doc(`
					Sends a test email to the administrator
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *TestEmailCommand) execute() (r_err error) {
	senderToken, senderTokenErr := c.getEmailSenderToken()
	if senderTokenErr != nil {
		return senderTokenErr
	}

	senderAddress, senderAddressErr := c.getEmailSenderAddress()
	if senderAddressErr != nil {
		return senderAddressErr
	}

	receiverAddress, receiverAddressErr := c.getEmailReceiverAddress()
	if receiverAddressErr != nil {
		return receiverAddressErr
	}

	to := []string{receiverAddress}

	// Message
	message := []byte("Subject: Hello from Go!\r\n" +
		"\r\n" +
		"This is a test email sent from a Go program.\r\n")

	// SMTP server config
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Auth
	auth := smtp.PlainAuth("", senderAddress, senderToken, smtpHost)

	// Send email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderAddress, to, message)
	if err != nil {
		c.PrintErrorf("Failed to send email")
		return err
	}

	c.Printf("Email sent successfully!")
	return nil
}

func (c *TestEmailCommand) getEmailSenderToken() (string, error) {
	return c.GetConfigurationString(common.ConfigKeyEmailSenderToken)
}

func (c *TestEmailCommand) getEmailSenderAddress() (string, error) {
	return c.GetConfigurationString(common.ConfigKeyEmailSenderAddress)
}

func (c *TestEmailCommand) getEmailReceiverAddress() (string, error) {
	return c.GetConfigurationString(common.ConfigKeyEmailReceiverAddress)
}
