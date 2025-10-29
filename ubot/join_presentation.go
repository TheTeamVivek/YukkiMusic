package ubot

import (
	"slices"

	tg "github.com/amarnathcjd/gogram/telegram"

	"github.com/TheTeamVivek/YukkiMusic/ntgcalls"
)

func (ctx *Context) joinPresentation(chatId int64, join bool) error {
	defer func() {
		if ctx.waitConnect[chatId] != nil {
			delete(ctx.waitConnect, chatId)
		}
	}()
	connectionMode, err := ctx.binding.GetConnectionMode(chatId)
	if err != nil {
		return err
	}
	if connectionMode == ntgcalls.StreamConnection {
		if ctx.pendingConnections[chatId] != nil {
			ctx.pendingConnections[chatId].Presentation = join
		}
	} else if connectionMode == ntgcalls.RtcConnection {
		if join {
			if !slices.Contains(ctx.presentations, chatId) {
				ctx.waitConnect[chatId] = make(chan error)
				jsonParams, err := ctx.binding.InitPresentation(chatId)
				if err != nil {
					return err
				}
				resultParams := "{\"transport\": null}"
				callResRaw, err := ctx.app.PhoneJoinGroupCallPresentation(
					ctx.inputGroupCalls[chatId],
					&tg.DataJson{
						Data: jsonParams,
					},
				)
				if err != nil {
					return err
				}
				callRes := callResRaw.(*tg.UpdatesObj)
				for _, u := range callRes.Updates {
					switch update := u.(type) {
					case *tg.UpdateGroupCallConnection:
						resultParams = update.Params.Data
					}
				}
				err = ctx.binding.Connect(
					chatId,
					resultParams,
					true,
				)
				if err != nil {
					return err
				}
				<-ctx.waitConnect[chatId]
				ctx.presentations = append(ctx.presentations, chatId)
			}
		} else if slices.Contains(ctx.presentations, chatId) {
			ctx.presentations = stdRemove(ctx.presentations, chatId)
			err = ctx.binding.StopPresentation(chatId)
			if err != nil {
				return err
			}
			_, err = ctx.app.PhoneLeaveGroupCallPresentation(
				ctx.inputGroupCalls[chatId],
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
