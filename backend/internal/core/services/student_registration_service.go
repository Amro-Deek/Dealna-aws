package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

type StudentRegistrationService struct {
	users      ports.IUserRepository
	preRegs    ports.IStudentPreRegistrationRepository
	email      ports.IEmailService
	hasher     ports.IPasswordHasher
	university ports.IUniversityRepository
}

func NewStudentRegistrationService(
	users ports.IUserRepository,
	preRegs ports.IStudentPreRegistrationRepository,
	email ports.IEmailService,
	hasher ports.IPasswordHasher,
	university ports.IUniversityRepository,
) *StudentRegistrationService {
	return &StudentRegistrationService{
		users:      users,
		preRegs:    preRegs,
		email:      email,
		hasher:     hasher,
		university: university,
	}
}
func (s *StudentRegistrationService) isValidUniversityEmail(
	ctx context.Context,
	email string,
) bool {

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := strings.ToLower(parts[1])

	// Only allow student domain
	if domain != "student.birzeit.edu" {
		return false
	}

	uni, err := s.university.GetByDomain(ctx, "birzeit.edu")
	if err != nil {
		return false
	}

	return uni.Status == "ACTIVE"
}

func (s *StudentRegistrationService) RequestStudentActivation(
	ctx context.Context,
	email string,
) error {

	if !s.isValidUniversityEmail(ctx, email) {
		return middleware.NewValidationError("email", "invalid university domain")
	}

	// check already user
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return middleware.NewEmailAlreadyUsedError(email)
	}

	token := uuid.NewString()

	pre := &domain.StudentPreRegistration{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.preRegs.Create(ctx, pre); err != nil {
		return middleware.NewDatabaseError("create prereg", err)
	}

	return s.email.SendActivationLink(email, token)
}
func (s *StudentRegistrationService) VerifyActivation(
	ctx context.Context,
	token string,
) error {

	pre, err := s.preRegs.GetByToken(ctx, token)
	if err != nil {
		return middleware.NewUnauthorizedError("invalid token")
	}

	if pre.VerifiedAt != nil {
		return middleware.NewUnauthorizedError("already verified")
	}

	if time.Now().After(pre.ExpiresAt) {
		return middleware.NewUnauthorizedError("token expired")
	}

	now := time.Now()
	pre.VerifiedAt = &now

	if err := s.preRegs.Update(ctx, pre); err != nil {
		return middleware.NewDatabaseError("update prereg", err)
	}

	return nil
}
func (s *StudentRegistrationService) CompleteStudentRegistration(
	ctx context.Context,
	email string,
	displayName string,
	password string,
	major *string,
	year *int,
) error {

	pre, err := s.preRegs.GetByEmail(ctx, email)
	if err != nil {
		return middleware.NewUnauthorizedError("activation not requested")
	}

	// لازم يكون verified من رابط الايميل
	if pre.VerifiedAt == nil {
		return middleware.NewUnauthorizedError("email not verified")
	}

	if pre.UsedAt != nil {
		return middleware.NewUnauthorizedError("registration already completed")
	}

	if time.Now().After(pre.ExpiresAt) {
		return middleware.NewUnauthorizedError("activation expired")
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return middleware.NewInternalError(err)
	}

	// ✅ domain check
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return middleware.NewValidationError("email", "invalid email")
	}

	if parts[1] != "student.birzeit.edu" {
		return middleware.NewValidationError("email", "only birzeit student emails allowed")
	}

	// ✅ university
	uni, err := s.university.GetByDomain(ctx, "birzeit.edu")
	if err != nil {
		return middleware.NewDatabaseError("get university", err)
	}

	// ✅ student id
	studentID := extractStudentID(email)

	_, err = s.users.CreateStudent(
		ctx,
		displayName,
		email,
		hash,
		major,
		year,
		uni.ID,
		studentID,
	)
	if err != nil {
		return middleware.NewDatabaseError("create user", err)
	}

	now := time.Now()
	pre.UsedAt = &now

	if err := s.preRegs.Update(ctx, pre); err != nil {
		return middleware.NewDatabaseError("update prereg", err)
	}

	return nil
}





func (s *StudentRegistrationService) ResendActivation(
	ctx context.Context,
	email string,
) error {

	pre, err := s.preRegs.GetByEmail(ctx, email)
	if err != nil {
		return middleware.NewUnauthorizedError("no pending activation")
	}

	// check resend limit
	if pre.ResendCount >= 3 &&
		pre.ResendWindowStart != nil &&
		time.Since(*pre.ResendWindowStart) < 30*time.Minute {

		return middleware.NewValidationError("email", "resend limit exceeded")
	}

	newToken := uuid.NewString()
	exp := time.Now().Add(24 * time.Hour)

	now := time.Now()

	if pre.ResendWindowStart == nil || time.Since(*pre.ResendWindowStart) > 30*time.Minute {
		pre.ResendCount = 0
		pre.ResendWindowStart = &now
	}

	pre.ResendCount++
	pre.Token = newToken
	pre.ExpiresAt = exp

	if err := s.preRegs.Update(ctx, pre); err != nil {
		return middleware.NewDatabaseError("update prereg", err)
	}

	return s.email.SendActivationLink(email, newToken)
}


func extractStudentID(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0] // numbers before @
}
