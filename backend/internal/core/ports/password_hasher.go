package ports

type IPasswordHasher interface {
    Compare(hash string, password string) error
}