package store

import (
	"github.com/nedpals/supabase-go"
)

func NewSupabaseClient(supabaseURL, supabaseKey string) *supabase.Client {
	return supabase.CreateClient(supabaseURL, supabaseKey)
}
