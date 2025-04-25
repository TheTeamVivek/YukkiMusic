# TODO

## 1.

```python

                    if popped.get("mystic"):
                        try:
                            await popped.get("mystic").delete()
                        except Exception:
                            pass
```

Merge this in Cleanmode That delete the message with **CLEANMODE_DELETE_MINS**

## 2.

Implement a env var or using db that help we can disable the Cleanmode and enable it


## 3.
  on M3u8 mode Track return a medium quality audio + video url add support for seperate high qulaity audio and video

## 4.
 
  On livestream every time returns audio url thats cause video playback not work add support for videp by returing or processing the audio and video url
  
## 5. Add this yml files and amd thier translations

```yml

call_11: "**No Active Voice Chat Found**\n\nPlease make sure group's voice chat is enabled. If already enabled, please end it and start fresh voice chat again and if the problem continues, try /reboot"
call_12: "**ASSISTANT IS ALREADY IN VOICECHAT** \n\nMusic bot system detected that assistant is already in the voicechat, if the problem continues restart the videochat and try again."
call_13: "**TELEGRAM SERVER ERROR**\n\nPlease restart Your voicechat."

tg_3: "**{0} Telagram Media Downloader**\n\n**Total file size:** {1}\n**Completed:** {2} \n**Percentage:** {3}%\n\n**Speed:** {4}/s\n**Elapsed Time:** {5}"
tg_4: "Sucessfully Downloaded\n Processing File Now..."
tg_5: "Download Already Completed."
tg_6: "Downloading already Cancelled."

enable: "enable"
disable: "disable"

spotify_1: "This Bot can't play spotify tracks and playlist, please contact my owner and ask him to add Spotify player."

ac_1: "Getting Active Voicechats....\nPlease hold on"
ac_2: "No active Chats Found"
ac_3: "**Active Voice Chat's:-**\n\n"
ac_4: "Active Chats info:\nAudio: {0}\nVideo: {1}"

anon_admin: "How to Fix this?"
anon_admin2: "You are an anonymous admin\nRevert back to user to use me"

TG_B_1: "🚦 Cancel downloading"

VPLAY_FLAGS: ["-v"]
CPLAY_FLAGS: ["-c"]
FPLAY_FLAGS: ["-f", "-force"]


logger_text: |
**{bot_mention} Play Log**

**Chat ID:** `{chat_id}`
**Chat Name:** {title}
**Chat Username:** {chatusername}

**User ID:** `{sender_id}`
**Name:** {user_mention}
**Username:** {username}

**Query:** {query}
**Stream Type:** {streamtype}

lang_1: "You are already using same language"
lang_2: "Your language changed successfully.."
lang_3: "Failed to change language or language in under Upadte"

admin_21 : "❌ Failed to reload admincache make sure bot is an admin in your chat"
```

## 6.
  If Possible so remove the /gstats,  it uses userdb and chattopdb that consumes too Database and still useless