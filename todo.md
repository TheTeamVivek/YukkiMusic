# 1.
```python
                    if popped.get("mystic"):
                        try:
                            await popped.get("mystic").delete()
                        except Exception:
                            pass
```

Merge this in Cleanmode That delete the message with **CLEANMODE_DELETE_MINS**

# 2.

Implement a env var or using db that help we can disable the Cleanmode and enable it