package filters

import "llm-router/types"

// GetSystemDefaultGroups returns a set of pre-tuned intent groups
func GetSystemDefaultGroups() []types.SemanticGroup {
	return []types.SemanticGroup{
		{
			Name:              "coding",
			IntentDescription: "Software development, programming, and technical problem solving.",
			Examples: []string{
				"write a binary search algorithm in python",
				"how do I fix a null pointer exception in java?",
				"refactor this function to be more efficient",
				"write a unit test for this go struct",
				"implement a linked list in c++",
				"explain this code snippet to me",
				"generate a regex for email validation",
			},
			RequiredCapability: "coding",
		},
		{
			Name:              "fast-chat",
			IntentDescription: "Simple conversational interactions, chit-chat, and basic knowledge questions.",
			Examples: []string{
				"hello! how is your day going?",
				"tell me a joke about robots",
				"what is the capital of france?",
				"how cold is it in london today?",
				"who won the world cup in 2022?",
				"translate 'hello' to spanish",
				"what is 2 + 2?",
			},
			AllowProviders: []string{"gemini"},
		},
	}
}
