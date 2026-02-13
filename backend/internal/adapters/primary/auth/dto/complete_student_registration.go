package dto

type CompleteStudentRegistrationRequest struct {
	Email        string  `json:"email"`
	DisplayName  string  `json:"displayName"`
	Password     string  `json:"password"`
	Major        *string `json:"major,omitempty"`
	AcademicYear *int    `json:"academicYear,omitempty"`
}

type CompleteStudentRegistrationResponse struct {
	Message string `json:"message"`
}
