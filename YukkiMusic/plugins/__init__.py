#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#
import glob
import os
import logging
import importlib
from os.path import dirname, isfile, join, abspath
import subprocess
from config import EXTRA_PLUGINS, EXTRA_PLUGINS_REPO, EXTRA_PLUGINS_FOLDER

# Define the path to the external plugins directory in the root of repo-a
ROOT_DIR = abspath(join(dirname(__file__), '..', '..'))
EXTERNAL_REPO_PATH = join(ROOT_DIR, EXTRA_PLUGINS_FOLDER)  # Local directory to clone the external repo

# Convert EXTRA_PLUGINS to a boolean
extra_plugins_enabled = EXTRA_PLUGINS.lower() == "true"

if extra_plugins_enabled:
    # Clone the external repository if not already cloned
    if not os.path.exists(EXTERNAL_REPO_PATH):
        subprocess.run(['git', 'clone', EXTRA_PLUGINS_REPO, EXTERNAL_REPO_PATH])

    # Install requirements if requirements.txt exists in the external plugins directory
    requirements_path = join(EXTERNAL_REPO_PATH, 'requirements.txt')
    if os.path.isfile(requirements_path):
        with open(os.devnull, 'w') as devnull:
            result = subprocess.run(['pip', 'install', '-r', requirements_path], stdout=devnull, stderr=subprocess.PIPE)
            if result.returncode != 0:
                logging.exception(result.stderr.decode())

def __list_all_modules():
    main_repo_plugins_dir = dirname(__file__)
    work_dirs = [main_repo_plugins_dir]

    if extra_plugins_enabled:
        work_dirs.append(EXTERNAL_REPO_PATH)

    all_modules = []

    for work_dir in work_dirs:
        mod_paths = glob.glob(join(work_dir, "*.py"))
        mod_paths += glob.glob(join(work_dir, "*/*.py"))
        
        modules = [
            (((f.replace(main_repo_plugins_dir, "YukkiMusic.plugins")).replace(EXTERNAL_REPO_PATH, EXTRA_PLUGINS_FOLDER)).replace(os.sep, "."))[:-3]
            for f in mod_paths
            if isfile(f) and f.endswith(".py") and not f.endswith("__init__.py")
        ]
        all_modules.extend(modules)

    return all_modules

ALL_MODULES = sorted(__list_all_modules())
__all__ = ALL_MODULES + ["ALL_MODULES"]