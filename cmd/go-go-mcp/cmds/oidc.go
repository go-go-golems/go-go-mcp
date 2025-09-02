package cmds

import (
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/auth/oidc"
	"github.com/spf13/cobra"
)

// NewOIDCCommand returns the root group for OIDC admin verbs.
func NewOIDCCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oidc",
		Short: "Manage embedded OIDC (users, tokens, clients)",
	}

	cmd.AddCommand(newOIDCUsersCommand())
	cmd.AddCommand(newOIDCTokensCommand())
	cmd.AddCommand(newOIDCClientsCommand())

	return cmd
}

func newOIDCUsersCommand() *cobra.Command {
	var db string
	users := &cobra.Command{
		Use:   "users",
		Short: "Manage users in SQLite",
	}
	users.PersistentFlags().StringVar(&db, "db", "", "SQLite DB path (required)")
	_ = users.MarkPersistentFlagRequired("db")

	// add
	var username, password string
	add := &cobra.Command{
		Use:   "add",
		Short: "Create user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := oidc.CreateUserInDB(db, username, password); err != nil {
				return err
			}
			fmt.Printf("User %s created\n", username)
			return nil
		},
	}
	add.Flags().StringVar(&username, "username", "", "Username")
	add.Flags().StringVar(&password, "password", "", "Password")
	_ = add.MarkFlagRequired("username")
	_ = add.MarkFlagRequired("password")
	users.AddCommand(add)

	// passwd
	username = ""
	password = ""
	passwd := &cobra.Command{
		Use:   "passwd",
		Short: "Set user password",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := oidc.SetUserPasswordInDB(db, username, password); err != nil {
				return err
			}
			fmt.Printf("Password updated for %s\n", username)
			return nil
		},
	}
	passwd.Flags().StringVar(&username, "username", "", "Username")
	passwd.Flags().StringVar(&password, "password", "", "New password")
	_ = passwd.MarkFlagRequired("username")
	_ = passwd.MarkFlagRequired("password")
	users.AddCommand(passwd)

	// del
	username = ""
	del := &cobra.Command{
		Use:   "del",
		Short: "Delete user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := oidc.DeleteUserInDB(db, username); err != nil {
				return err
			}
			fmt.Printf("User %s deleted\n", username)
			return nil
		},
	}
	del.Flags().StringVar(&username, "username", "", "Username")
	_ = del.MarkFlagRequired("username")
	users.AddCommand(del)

	// list
	list := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE: func(cmd *cobra.Command, args []string) error {
			us, err := oidc.ListUsersInDB(db)
			if err != nil {
				return err
			}
			for _, u := range us {
				fmt.Printf("%s\tdisabled=%t\tcreated=%s\n", u.Username, u.Disabled, u.CreatedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
	users.AddCommand(list)

	return users
}

func newOIDCTokensCommand() *cobra.Command {
	var db string
	var token string
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Manage dev tokens in SQLite",
	}
	cmd.PersistentFlags().StringVar(&db, "db", "", "SQLite DB path (required)")
	_ = cmd.MarkPersistentFlagRequired("db")

	list := &cobra.Command{
		Use:   "list",
		Short: "List tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			ts, err := oidc.ListTokensInDB(db)
			if err != nil {
				return err
			}
			for _, t := range ts {
				fmt.Printf("%s\tsub=%s\tclient=%s\texpires=%s\t%s\n", t.Token, t.Subject, t.ClientID, t.ExpiresAt.Format("2006-01-02 15:04:05"), t.Scopes)
			}
			return nil
		},
	}
	cmd.AddCommand(list)

	del := &cobra.Command{
		Use:   "del",
		Short: "Delete token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return oidc.DeleteTokenInDB(db, token)
		},
	}
	del.Flags().StringVar(&token, "token", "", "Token value")
	_ = del.MarkFlagRequired("token")
	cmd.AddCommand(del)

	return cmd
}

func newOIDCClientsCommand() *cobra.Command {
	var db string
	var id string
	var redirects []string
	cmd := &cobra.Command{
		Use:   "clients",
		Short: "Manage OAuth clients in SQLite",
	}
	cmd.PersistentFlags().StringVar(&db, "db", "", "SQLite DB path (required)")
	_ = cmd.MarkPersistentFlagRequired("db")

	list := &cobra.Command{
		Use:   "list",
		Short: "List clients",
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := oidc.ListClientsInDB(db)
			if err != nil {
				return err
			}
			for _, c := range cs {
				fmt.Printf("%s\t%v\n", c.ClientID, c.RedirectURIs)
			}
			return nil
		},
	}
	cmd.AddCommand(list)

	upsert := &cobra.Command{
		Use:   "upsert",
		Short: "Create or update an OAuth client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return oidc.PersistClientInDB(db, id, redirects)
		},
	}
	upsert.Flags().StringVar(&id, "id", "", "Client ID")
	upsert.Flags().StringArrayVar(&redirects, "redirect-uri", []string{}, "Redirect URIs (repeat) or comma-separated")
	_ = upsert.MarkFlagRequired("id")
	cmd.AddCommand(upsert)

	return cmd
}
