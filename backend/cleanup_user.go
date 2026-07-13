package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/Nerzal/gocloak/v13"
)

func main() {
	email := "1221645@student.birzeit.edu"
	dbURL := "postgres://dealna_user:amro123@98.92.82.224:5432/dealna_db?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("DB error:", err)
		return
	}
	defer db.Close()

	kc := gocloak.NewClient("http://54.197.49.30:8080")
	ctx := context.Background()
	token, err := kc.LoginClient(ctx, "dealna-backend", "HtcTRtda0F53HXbf3uIUpTqakX2albXQ", "Dealna")
	if err != nil {
		fmt.Println("KC login err:", err)
		return
	}

	fmt.Println("=== Deleting user:", email, "===")
	// Get user_id
	var userID string
	err = db.QueryRow(`SELECT user_id FROM public."User" WHERE email = $1 LIMIT 1`, email).Scan(&userID)
	if err == sql.ErrNoRows {
		fmt.Println("User not found in DB")
	} else if err != nil {
		fmt.Println("Error finding user:", err)
	} else {
		// Delete related data first
		_, _ = db.Exec(`DELETE FROM provider_application_document WHERE application_id IN (SELECT id FROM provider_application WHERE applicant_id = $1)`, userID)
		_, _ = db.Exec(`DELETE FROM provider_application WHERE applicant_id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM public.item WHERE owner_id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM provider WHERE user_id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM student WHERE user_id = $1`, userID) // Added this for student accounts!
		_, _ = db.Exec(`DELETE FROM profile WHERE user_id = $1`, userID)
		_, err = db.Exec(`DELETE FROM public."User" WHERE user_id = $1`, userID)
		if err != nil { fmt.Println("Error deleting User:", err) } else { fmt.Println("Deleted User from DB") }
	}

	// Delete pre_reg
	_, err = db.Exec(`DELETE FROM provider_pre_registration WHERE email = $1`, email)
	if err != nil { fmt.Println("Error deleting provider_pre_registration:", err) }
	_, err = db.Exec(`DELETE FROM student_pre_registration WHERE email = $1`, email)
	if err != nil { fmt.Println("Error deleting student_pre_registration:", err) }

	fmt.Println("DB cleanup done")
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
