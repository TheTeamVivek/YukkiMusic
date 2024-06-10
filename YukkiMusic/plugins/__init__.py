#
# Copyright (C) 2024 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#
import glob
import os
import importlib
from os.path import dirname, isfile, join, abspath
import subprocess
import logging
import sys
from config import EXTRA_PLUGINS, EXTRA_PLUGINS_REPO, EXTRA_PLUGINS_FOLDER

# Set up logging
logging.basicConfig(level=logging.ERROR)
logger = logging.getLogger(__name__)

# Define the path to the external plugins directory in the root of repo-a
ROOT_DIR = abspath(join(dirname(__file__), '..', '..'))
EXTERNAL_REPO_PATH = join(ROOT_DIR, EXTRA_PLUGINS_FOLDER)  # Local directory to clone the external repo

# Convert EXTRA_PLUGINS to a boolean
extra_plugins_enabled = EXTRA_PLUGINS.lower() == "true"

if extra_plugins_enabled:
    # Clone the external repository if not already cloned
    if not os.path.exists(EXTERNAL_REPO_PATH):
        with open(os.devnull, 'w') as devnull:
            clone_result = subprocess.run(
                ['git', 'clone', EXTRA_PLUGINS_REPO, EXTERNAL_REPO_PATH],
                stdout=devnull,
                stderr=subprocess.PIPE
            )
            if clone_result.returncode != 0:
                logger.error(f"Error cloning external plugins repository: {clone_result.stderr.decode()}")

    # Check if utils folder exists in the external repo
    utils_path = join(EXTERNAL_REPO_PATH, 'utils')
    if os.path.isdir(utils_path):
        # Add the utils folder path to sys.path if it exists
        sys.path.append(utils_path)

    # Install requirements if requirements.txt exists in the external plugins directory
    requirements_path = join(EXTERNAL_REPO_PATH, 'requirements.txt')
    if os.path.isfile(requirements_path):
        with open(os.devnull, 'w') as devnull:
            install_result = subprocess.run(
                ['pip', 'install', '-r', requirements_path],
                stdout=devnull,
                stderr=subprocess.PIPE
            )
            if install_result.returncode != 0:
                logger.error(f"Error installing requirements for external plugins: {install_result.stderr.decode()}")

def __list_all_modules():
    # Define directories to search for plugins
    main_repo_plugins_dir = dirname(__file__)
    work_dirs = [main_repo_plugins_dir]

    if extra_plugins_enabled:
        work_dirs.append(join(EXTERNAL_REPO_PATH, 'plugins'))

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