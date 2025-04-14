package domain

import "time"

// User represents a user entity in the e-wallet system
type User struct {
	ID        int        `json:"id"`
	Fullname  string     `json:"fullname"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
)

// NewUser creates a new user entity
// using constructor pattern
func NewUser(fullname, email, phone string) *User {
	now := time.Now()
	return &User{
		Fullname:  fullname,
		Email:     email,
		Phone:     phone,
		Status:    UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Notes
// What if we have so many params on a func?
// ans:
// we can create new struct params and pass it to a func
// e.g
// type UserParams struct {
// 	Fullname    string
// 	Email       string
// 	Phone       string
// 	Address     string
// 	DateOfBirth time.Time
// 	Preferences map[string]interface{}
// 	... many more fields
// }
//
// func NewUser(params UserParams) *User {
// 	now := time.Now()
// 	return &User{
// 			Fullname:    params.Fullname,
// 			Email:       params.Email,
// 			Phone:       params.Phone,
// 			Address:     params.Address,
// 			DateOfBirth: params.DateOfBirth,
// 			Preferences: params.Preferences,
// 			Status:      "active",
// 			CreatedAt:   now,
// 			UpdatedAt:   now,
// 	}
// }
// 2. In Ports and Adapter Architecture, I often found mapping functions pattern
// between core domain with Infracture DTOs and high risk of code duplication.
// So what should we do?
// ans:
// if I had to choose between strictly adhering to hexagonal archicture principles
// versus reducing mapping functions and code duplication, I would make a pragmatic
// decision based on projects size.
// for smaller to medium-sized project with a stable domain and relatively simple
// persistence needs, I would like opt for practical compromise of combining tags
// e.g type ID int `json:"id"; db:"fullname"`.
// for large project, I would strict to hexagonal architecture rule.
