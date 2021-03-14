package eod

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type hintCombo struct {
	exists int
	text   string
}

func (b *EoD) hintCmd(elem string, m msg, rsp rsp) {
	lock.RLock()
	dat, exists := b.dat[m.GuildID]
	lock.RUnlock()
	if !exists {
		return
	}
	el, exists := dat.elemCache[strings.ToLower(elem)]
	if !exists {
		rsp.ErrorMessage(fmt.Sprintf("Element %s doesn't exist!", elem))
	}
	inv, exists := dat.invCache[m.Author.ID]
	if !exists {
		rsp.ErrorMessage("You don't have an inventory!")
		return
	}

	combs, err := b.db.Query("SELECT elem1, elem2 FROM eod_combos WHERE elem3=? AND guild=?", elem, m.GuildID)
	if rsp.Error(err) {
		return
	}
	defer combs.Close()
	var elem1 string
	var elem2 string
	out := make([]hintCombo, 0)
	for combs.Next() {
		err = combs.Scan(&elem1, &elem2)
		if rsp.Error(err) {
			return
		}

		_, haselem1 := inv[strings.ToLower(elem1)]
		_, haselem2 := inv[strings.ToLower(elem2)]
		pref := x
		ex := 0
		if haselem1 && haselem2 {
			pref = check
			ex = 1
		}
		txt := fmt.Sprintf("%s %s + %s", pref, elem1, elem2)
		out = append(out, hintCombo{
			exists: ex,
			text:   txt,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].exists > out[j].exists
	})

	text := ""
	for _, val := range out {
		text += val.text + "\n"
	}

	rsp.Embed(&discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Hints for %s", el.Name),
		Description: text,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: el.Image,
		},
	})
}
