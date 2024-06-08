#
# Copyright (C) 2024-present by TeamYukki@Github, < https://github.com/TeamYukki >.
#
# This file is part of < https://github.com/TeamYukki/YukkiMusicBot > project,
# and is released under the "GNU v3.0 License Agreement".
# Please see < https://github.com/TeamYukki/YukkiMusicBot/blob/master/LICENSE >
#
# All rights reserved.
#
'''
import glob
from os.path import dirname, isfile


def __list_all_modules():
    work_dir = dirname(__file__)
    mod_paths = glob.glob(work_dir + "/*/*.py")

    all_modules = [
        (((f.replace(work_dir, "")).replace("/", "."))[:-3])
        for f in mod_paths
        if isfile(f) and f.endswith(".py") and not f.endswith("__init__.py")
    ]

    return all_modules


ALL_MODULES = sorted(__list_all_modules())
__all__ = ALL_MODULES + ["ALL_MODULES"]
'''


import glob
import os
import importlib
from os.path import dirname, isfile, join, abspath
import subprocess

# Step 1: Define the URL of the external repository
EXTERNAL_REPO_URL = 'https://github.com/user/repo-b.git'
# Step 2: Define the path to the plugins directory in the root of repo-a
ROOT_DIR = abspath(join(dirname(__file__), '..', '..'))
EXTERNAL_REPO_PATH = join(ROOT_DIR, 'plugins')  # Local directory to clone the external repo

# Step 3: Clone the external repository if not already cloned
if not os.path.exists(EXTERNAL_REPO_PATH):
    subprocess.run(['git', 'clone', EXTERNAL_REPO_URL, EXTERNAL_REPO_PATH])

# Step 4: Install requirements if requirements.txt exists in the plugins directory
requirements_path = join(EXTERNAL_REPO_PATH, 'requirements.txt')
if os.path.isfile(requirements_path):
    subprocess.run(['pip', 'install', '-r', requirements_path])

def __list_all_modules():
    # Define directories to search for plugins
    main_repo_plugins_dir = dirname(__file__)
    work_dirs = [main_repo_plugins_dir, EXTERNAL_REPO_PATH]
    all_modules = []

    for work_dir in work_dirs:
        mod_paths = glob.glob(join(work_dir, "*.py"))
        mod_paths += glob.glob(join(work_dir, "*/*.py"))
        
        modules = [
            (((f.replace(main_repo_plugins_dir, "YukkiMusic.plugins")).replace(EXTERNAL_REPO_PATH, "plugins")).replace(os.sep, "."))[:-3]
            for f in mod_paths
            if isfile(f) and f.endswith(".py") and not f.endswith("__init__.py")
        ]
        all_modules.extend(modules)

    return all_modules

ALL_MODULES = sorted(__list_all_modules())
__all__ = ALL_MODULES + ["ALL_MODULES"]