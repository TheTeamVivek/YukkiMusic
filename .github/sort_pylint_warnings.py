import json


with open("pylint_warnings.json") as file:
    pylint_data = json.load(file)

pylint_sorted_by_type = sorted(pylint_data, key=lambda x: x["path"])

dont_appers = [
    "too-many-statements",
    "duplicate-code",
    "too-many-branches",
    "cyclic-import",
]

for item in pylint_sorted_by_type:
    if (
        any(item["symbol"].lower() == symbol.lower() for symbol in dont_appers)
        or item["module"] == ".github.sort_pylint_warnings"
    ):
        pylint_sorted_by_type.remove(item)  # Currently disable duplicate codes
# with open("pylint_warnings_sorted.json", "w") as file:
with open("pylint_warnings.json", "w") as file:
    json.dump(pylint_sorted_by_type, file, indent=4)
