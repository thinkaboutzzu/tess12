package helper

import (
	"github.com/bwmarrin/discordgo"
	"jkt48lab/model"
)

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func RemoveStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func ChunkSlicePM(slice []model.PMMessage, chunkSize int) [][]model.PMMessage {
	var chunks [][]model.PMMessage
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func ChunkSliceEmbeds(slice []*discordgo.MessageEmbed, chunkSize int) [][]*discordgo.MessageEmbed {
	var chunks [][]*discordgo.MessageEmbed
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func SortMapToSlicePMStats(ranks map[string]model.PMRanking, keys []string, max int) []model.PMRanking {
	var ranksSlice []model.PMRanking
	for i, k := range keys {
		if i >= max {
			break
		}
		ranksSlice = append(ranksSlice, ranks[k])
	}
	return ranksSlice
}
