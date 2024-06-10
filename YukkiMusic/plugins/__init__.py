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
from pathlib import Path
import subprocess
import logging
import sys
from config import EXTRA_PLUGINS, EXTRA_PLUGINS_REPO, EXTRA_PLUGINS_FOLDER

# Set up logging
logging.basicConfig(level=logging.DEBUG)  # Set to DEBUG for more detailed logs
logger = logging.getLogger(__name__)

# Define the path to the external plugins directory in the root of repo-a
ROOT_DIR = Path(__file__).resolve().parents[2]
EXTERNAL_REPO_PATH = ROOT_DIR / EXTRA_PLUGINS_FOLDER  # Local directory to clone the external repo

# Convert EXTRA_PLUGINS to a boolean
extra_plugins_enabled = EXTRA_PLUGINS.lower() == "true"

if extra_plugins_enabled:
    # Clone the external repository if not already cloned
    if not EXTERNAL_REPO_PATH.exists():
        with open(os.devnull, 'w') as devnull:
            clone_result = subprocess.run(
                ['git', 'clone', EXTRA_PLUGINS_REPO, str(EXTERNAL_REPO_PATH)],
                stdout=devnull,
                stderr=subprocess.PIPE
            )
            if clone_result.returncode != 0:
                logger.error(f"Error cloning external plugins repository: {clone_result.stderr.decode()}")
            else:
                logger.info(f"Successfully cloned {EXTRA_PLUGINS_REPO} into {EXTERNAL_REPO_PATH}")

    # Log the directory structure after cloning
    for path in EXTERNAL_REPO_PATH.rglob('*'):
        logger.debug(f"Cloned file/folder: {path}")

    # Check if utils folder exists in the external repo
    utils_path = EXTERNAL_REPO_PATH / 'utils'
    if utils_path.is_dir():
        # Add the utils folder path to sys.path if it exists
        sys.path.append(str(utils_path))
        logger.info(f"Added {utils_path} to sys.path")
        # Log the contents of the utils directory
        logger.debug(f"Contents of utils directory: {os.listdir(utils_path)}")
    else:
        logger.error(f"utils directory not found in {EXTERNAL_REPO_PATH}")

    # Install requirements if requirements.txt exists in the external plugins directory
    requirements_path = EXTERNAL_REPO_PATH / 'requirements.txt'
    if requirements_path.is_file():
        with open(os.devnull, 'w') as devnull:
            install_result = subprocess.run(
                ['pip', 'install', '-r', str(requirements_path)],
                stdout=devnull,
                stderr=subprocess.PIPE
            )
            if install_result.returncode != 0:
                logger.error(f"Error installing requirements for external plugins: {install_result.stderr.decode()}")
            else:
                logger.info(f"Successfully installed requirements from {requirements_path}")

# Test import to check if utils can be imported
try:
    import utils.capture_err
    logger.info("Successfully imported utils.capture_err")
except ImportError as e:
    logger.error(f"Error importing utils: {e}")

def __list_all_modules():
    # Define directories to search for plugins
    main_repo_plugins_dir = Path(__file__).parent
    work_dirs = [main_repo_plugins_dir]

    if extra_plugins_enabled:
        work_dirs.append(EXTERNAL_REPO_PATH)

    all_modules = []

    for work_dir in work_dirs:
        mod_paths = glob.glob(str(work_dir / "*.py"))
        mod_paths += glob.glob(str(work_dir / "*/*.py"))
        
        modules = [
            (((f.replace(str(main_repo_plugins_dir), "YukkiMusic.plugins")).replace(str(EXTERNAL_REPO_PATH), EXTRA_PLUGINS_FOLDER)).replace(os.sep, "."))[:-3]
            for f in mod_paths
            if Path(f).is_file() and f.endswith(".py") and not f.endswith("__init__.py")
        ]
        all_modules.extend(modules)

    return all_modules

ALL_MODULES = sorted(__list_all_modules())
__all__ = ALL_MODULES + ["ALL_MODULES"]

# Additional debug log to print sys.path
logger.debug(f"sys.path: {sys.path}")
