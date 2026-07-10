package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/Nerzal/gocloak/v13"
)

func main() {
	email := "odehdeek187@gmail.com"
	dbURL := "postgres://dealna_user:amro123@98.92.82.224:5432/dealna_db?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("DB error:", err)
		return
	}
	defer db.Close()

	// Get user_id
	var userID string
	err = db.QueryRow(`SELECT user_id FROM public."User" WHERE email = $1 LIMIT 1`, email).Scan(&userID)
	if err == sql.ErrNoRows {
		fmt.Println("User not found in DB")
	} else if err != nil {
		fmt.Println("Error finding user:", err)
	} else {
		// Delete provider application documents
		_, err = db.Exec(`DELETE FROM provider_application_document WHERE application_id IN (SELECT id FROM provider_application WHERE applicant_id = $1)`, userID)
		if err != nil { fmt.Println("Error deleting provider_application_document:", err) }

		// Delete provider application
		_, err = db.Exec(`DELETE FROM provider_application WHERE applicant_id = $1`, userID)
		if err != nil { fmt.Println("Error deleting provider_application:", err) }

		// Delete items
		_, err = db.Exec(`DELETE FROM public.item WHERE owner_id = $1`, userID)
		if err != nil { fmt.Println("Error deleting items:", err) }

		// Delete provider
		_, err = db.Exec(`DELETE FROM provider WHERE user_id = $1`, userID)
		if err != nil { fmt.Println("Error deleting provider:", err) }

		// Delete User
		_, err = db.Exec(`DELETE FROM public."User" WHERE user_id = $1`, userID)
		if err != nil { fmt.Println("Error deleting User:", err) }
	}

	// Delete pre_reg
	_, err = db.Exec(`DELETE FROM provider_pre_registration WHERE email = $1`, email)
	if err != nil { fmt.Println("Error deleting provider_pre_registration:", err) }

	fmt.Println("DB cleanup done")

	// Keycloak
	kc := gocloak.NewClient("http://54.197.49.30:8080")
	ctx := context.Background()
	token, err := kc.LoginClient(ctx, "dealna-backend", "HtcTRtda0F53HXbf3uIUpTqakX2albXQ", "Dealna")
	if err != nil {
		fmt.Println("KC login err:", err)
		return
	}
	users, err := kc.GetUsers(ctx, token.AccessToken, "Dealna", gocloak.GetUsersParams{
		Email: gocloak.StringP(email),
		Exact: gocloak.BoolP(true),
	})
	if err != nil {
		fmt.Println("KC get err:", err)
		return
	}
	for _, u := range users {
		err = kc.DeleteUser(ctx, token.AccessToken, "Dealna", *u.ID)
		if err != nil {
			fmt.Println("KC delete err:", err)
		} else {
			fmt.Println("Deleted user from KC:", *u.ID)
		}
	}
}
