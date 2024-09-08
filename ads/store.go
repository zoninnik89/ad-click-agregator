package main

import "context"

type Store struct {
	// add db later
}

func NewStore() *Store {
	return &Store{}
}

func (store *Store) Create(context.Context) error {
	return nil
}
