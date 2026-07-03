package learning

import (
	"context"
	"strings"

	"glasses-english-ai/internal/domain"
)

type StaticDictionary struct {
	cards map[string]domain.LearningCard
}

func NewStaticDictionary() *StaticDictionary {
	cards := []domain.LearningCard{
		{English: "cup", Chinese: "杯子", Phonetic: "/kʌp/", ExampleSentence: "This is a cup.", ExampleMeaning: "这是一个杯子。"},
		{English: "book", Chinese: "书", Phonetic: "/bʊk/", ExampleSentence: "This is a book.", ExampleMeaning: "这是一本书。"},
		{English: "chair", Chinese: "椅子", Phonetic: "/tʃer/", ExampleSentence: "This is a chair.", ExampleMeaning: "这是一把椅子。"},
		{English: "phone", Chinese: "手机", Phonetic: "/foʊn/", ExampleSentence: "This is a phone.", ExampleMeaning: "这是一部手机。"},
		{English: "table", Chinese: "桌子", Phonetic: "/ˈteɪbəl/", ExampleSentence: "This is a table.", ExampleMeaning: "这是一张桌子。"},
		{English: "pen", Chinese: "笔", Phonetic: "/pen/", ExampleSentence: "This is a pen.", ExampleMeaning: "这是一支笔。"},
	}

	dict := &StaticDictionary{cards: make(map[string]domain.LearningCard, len(cards))}
	for _, card := range cards {
		dict.cards[card.English] = card
	}
	return dict
}

func (d *StaticDictionary) FindCard(_ context.Context, english string) (domain.LearningCard, error) {
	key := strings.ToLower(strings.TrimSpace(english))
	if card, ok := d.cards[key]; ok {
		return card, nil
	}
	return domain.LearningCard{
		English:         key,
		Chinese:         "未知物体",
		Phonetic:        "",
		ExampleSentence: "This is a " + key + ".",
		ExampleMeaning:  "这是一个" + key + "。",
	}, nil
}
