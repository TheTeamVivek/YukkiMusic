# TODO NECESSARY

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
  
## 5. To Work better with multiple lanaguge as command we can Merge all command lanaguge command in files


## 6.
  If Possible so remove the /gstats,  it uses userdb and chattopdb that consumes too Database and still useless


## 7.

  Now we can say that YukkiMusic fully support multiples langauge with command, enable, disable so we need to refactor all languages