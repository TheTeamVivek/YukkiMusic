#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.

import asyncio
import logging

from git import Repo
from git.exc import GitCommandError, InvalidGitRepositoryError

import config

loop = asyncio.get_event_loop_policy().get_event_loop()
logger = logging.getLogger(__name__)


def install_req() -> tuple[str, str, int, int]:
    async def install_requirements():
        process = await asyncio.create_subprocess_shell(
            "uv pip install --system -e .",
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )
        stdout, stderr = await process.communicate()
        return (
            stdout.decode("utf-8", "replace").strip(),
            stderr.decode("utf-8", "replace").strip(),
            process.returncode,
            process.pid,
        )

    return loop.run_until_complete(install_requirements())


def git():
    if config.GIT_TOKEN:
        git_username = config.UPSTREAM_REPO.split("com/")[1].split("/")[0]
        temp_repo = config.UPSTREAM_REPO.split("https://")[1]
        upstream_repo = f"https://{git_username}:{config.GIT_TOKEN}@{temp_repo}"
    else:
        upstream_repo = config.UPSTREAM_REPO

    try:
        repo = Repo()
        logger.info("Git Client Found [VPS DEPLOYER]")
    except GitCommandError:
        logger.info("Invalid Git Command")
    except InvalidGitRepositoryError:
        repo = Repo.init()
        if "origin" in repo.remotes:
            origin = repo.remote("origin")
        else:
            origin = repo.create_remote("origin", upstream_repo)
        origin.fetch()
        repo.create_head(
            config.UPSTREAM_BRANCH,
            origin.refs[config.UPSTREAM_BRANCH],
        )
        repo.heads[config.UPSTREAM_BRANCH].set_tracking_branch(
            origin.refs[config.UPSTREAM_BRANCH]
        )
        repo.heads[config.UPSTREAM_BRANCH].checkout(True)

        try:
            repo.create_remote("origin", config.UPSTREAM_REPO)
        except Exception:
            pass

    nrs = repo.remote("origin")
    nrs.fetch(config.UPSTREAM_BRANCH)

    try:
        nrs.pull(config.UPSTREAM_BRANCH)
    except GitCommandError:
        repo.git.reset("--hard", "FETCH_HEAD")
    install_req()
    logger.info("Fetched Updates from: %s", config.UPSTREAM_REPO)
