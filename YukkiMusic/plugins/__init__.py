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
import importlib
import logging
import os
import shutil
import subprocess
import sys
from os.path import abspath, dirname, isfile, join

from YukkiMusic import LOGGER
from config import EXTRA_PLUGINS, EXTRA_PLUGINS_FOLDER, EXTRA_PLUGINS_REPO

logger = LOGGER(__name__)

ROOT_DIR = abspath(join(dirname(__file__), "..", ".."))
EXTERNAL_REPO_PATH = join(ROOT_DIR, EXTRA_PLUGINS_FOLDER)
extra_plugins_enabled = EXTRA_PLUGINS.lower() == "true"

if extra_plugins_enabled:
    if not os.path.exists(EXTERNAL_REPO_PATH):
        logger.info("Cloning external plugins repository for the first time...")
        with open(os.devnull, "w") as devnull:
            clone_result = subprocess.run(
                ["git", "clone", EXTRA_PLUGINS_REPO, EXTERNAL_REPO_PATH],
                stdout=devnull,
                stderr=subprocess.PIPE,
            )
            if clone_result.returncode != 0:
                logger.error(
                    f"Error cloning external plugins repository: {clone_result.stderr.decode()}"
                )
                logger.info(
                    "External plugins repository is private or inaccessible. Using previously saved plugins if available."
                )
                if not os.path.exists(EXTERNAL_REPO_PATH):
                    logger.critical(
                        "No previously saved external plugins found. Exiting."
                    )
                    sys.exit(1)
    else:
        logger.info("External plugins repository already exists. Pulling updates...")
        with open(os.devnull, "w") as devnull:
            pull_result = subprocess.run(
                ["git", "-C", EXTERNAL_REPO_PATH, "pull"],
                stdout=devnull,
                stderr=subprocess.PIPE,
            )
            if pull_result.returncode != 0:
                logger.error(
                    f"Error pulling updates from external plugins repository: {pull_result.stderr.decode()}"
                )
                logger.info("Using the previously saved plugins without updates.")
    # Copy utils folder if exists
    utils_source_path = join(EXTERNAL_REPO_PATH, "utils")
    utils_target_path = join(ROOT_DIR, "utils")
    if os.path.isdir(utils_source_path):
        if not os.path.exists(utils_target_path):
            os.rename(utils_source_path, utils_target_path)
        else:
            for root, dirs, files in os.walk(utils_source_path):
                relative_path = os.path.relpath(root, utils_source_path)
                target_dir = os.path.join(utils_target_path, relative_path)
                os.makedirs(target_dir, exist_ok=True)
                for file in files:
                    source_file = os.path.join(root, file)
                    target_file = os.path.join(target_dir, file)
                    if not os.path.exists(target_file):
                        os.rename(source_file, target_file)

    if os.path.isdir(utils_target_path):
        sys.path.append(utils_target_path)

    # Install requirements if present
    requirements_path = join(EXTERNAL_REPO_PATH, "requirements.txt")
    if os.path.isfile(requirements_path):
        with open(os.devnull, "w") as devnull:
            install_result = subprocess.run(
                ["pip", "install", "-r", requirements_path],
                stdout=devnull,
                stderr=subprocess.PIPE,
            )
            if install_result.returncode != 0:
                logger.error(
                    f"Error installing requirements for external plugins: {install_result.stderr.decode()}"
                )


def __list_all_modules():
    main_repo_plugins_dir = dirname(__file__)
    work_dirs = [main_repo_plugins_dir]

    if extra_plugins_enabled:
        logger.info("Loading extra plugins...")
        work_dirs.append(join(EXTERNAL_REPO_PATH, "plugins"))

    all_modules = []

    for work_dir in work_dirs:
        mod_paths = glob.glob(join(work_dir, "*.py"))
        mod_paths += glob.glob(join(work_dir, "*/*.py"))

        modules = [
            (
                (
                    (f.replace(main_repo_plugins_dir, "YukkiMusic.plugins")).replace(
                        EXTERNAL_REPO_PATH, EXTRA_PLUGINS_FOLDER
                    )
                ).replace(os.sep, ".")
            )[:-3]
            for f in mod_paths
            if isfile(f) and f.endswith(".py") and not f.endswith("__init__.py")
        ]
        all_modules.extend(modules)

    return all_modules


ALL_MODULES = sorted(__list_all_modules())
__all__ = ALL_MODULES + ["ALL_MODULES"]
