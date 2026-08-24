package common

import (
	"bctbackend/algorithms"
	"bctbackend/database/queries"
	"bctbackend/dictionary"
	"errors"
	"math/rand/v2"
)

var ErrOutOfPasswords = errors.New("out of passwords")

func FindUnusedPassword(db queries.DatabaseQuerier) (string, error) {
	passwords, err := queries.CollectPasswords(db)
	if err != nil {
		return "", err
	}
	passwordSet := algorithms.NewSet(passwords...)

	indices := algorithms.Range(0, len(passwords))
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	for _, index := range indices {
		candidatePassword := dictionary.Passwords[index]

		if !passwordSet.Contains(candidatePassword) {
			return candidatePassword, nil
		}
	}

	return "", ErrOutOfPasswords
}
