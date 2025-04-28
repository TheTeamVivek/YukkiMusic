#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from typing import Any

import aiohttp

__all__ = ["Request"]


class Request:
    @staticmethod
    async def get_json(
        url: str,
        params: dict[str, Any] | None = None,
        headers: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """
        Sends a GET request and returns the response as JSON.

        Args:
            url (str): The URL to send the GET request to.
            params (Optional[Dict[str, Any]]): Query parameters to include in the request. Defaults to None.
            headers (Optional[Dict[str, str]]): Headers to include in the request. Defaults to None.

        Returns:
            Dict[str, Any]: The JSON response from the request.
        """
        async with aiohttp.ClientSession() as session:
            async with session.get(url, params=params, headers=headers) as resp:
                return await resp.json()

    @staticmethod
    async def get_raw(
        url: str,
        params: dict[str, Any] | None = None,
        headers: dict[str, str] | None = None,
    ) -> bytes:
        """
        Sends a GET request and returns the raw response as bytes.

        Args:
            url (str): The URL to send the GET request to.
            params (Optional[Dict[str, Any]]): Query parameters to include in the request. Defaults to None.
            headers (Optional[Dict[str, str]]): Headers to include in the request. Defaults to None.

        Returns:
            bytes: The raw response from the request.
        """
        async with aiohttp.ClientSession() as session:
            async with session.get(url, params=params, headers=headers) as resp:
                return await resp.read()

    @staticmethod
    async def get_text(
        url: str,
        params: dict[str, Any] | None = None,
        headers: dict[str, str] | None = None,
    ) -> str:
        """
        Sends a GET request and returns the response as a string.

        Args:
            url (str): The URL to send the GET request to.
            params (Optional[Dict[str, Any]]): Query parameters to include in the request. Defaults to None.
            headers (Optional[Dict[str, str]]): Headers to include in the request. Defaults to None.

        Returns:
            str: The response as a string.
        """
        async with aiohttp.ClientSession() as session:
            async with session.get(url, params=params, headers=headers) as resp:
                return await resp.text()

    @staticmethod
    async def post_json(
        url: str,
        data: dict[str, Any] | Any | None = None,
        headers: dict[str, str] | None = None,
    ) -> dict[str, Any]:
        """
        Sends a POST request with JSON data and returns the response as JSON.

        Args:
            url (str): The URL to send the POST request to.
            data (Optional[Union[Dict[str, Any], Any]]): JSON data to include in the request body. Defaults to None.
            headers (Optional[Dict[str, str]]): Headers to include in the request. Defaults to None.

        Returns:
            Dict[str, Any]: The JSON response from the request.
        """
        async with aiohttp.ClientSession() as session:
            async with session.post(url, json=data, headers=headers) as resp:
                return await resp.json()

    @staticmethod
    async def post_raw(
        url: str,
        data: dict[str, Any] | Any | None = None,
        headers: dict[str, str] | None = None,
    ) -> bytes:
        """
        Sends a POST request with JSON data and returns the raw response as bytes.

        Args:
            url (str): The URL to send the POST request to.
            data (Optional[Union[Dict[str, Any], Any]]): JSON data to include in the request body. Defaults to None.
            headers (Optional[Dict[str, str]]): Headers to include in the request. Defaults to None.

        Returns:
            bytes: The raw response from the request.
        """
        async with aiohttp.ClientSession() as session:
            async with session.post(url, json=data, headers=headers) as resp:
                return await resp.read()

    @staticmethod
    async def post_text(
        url: str,
        data: dict[str, Any] | Any | None = None,
        headers: dict[str, str] | None = None,
    ) -> str:
        """
        Sends a POST request with JSON data and returns the response as a string.

        Args:
            url (str): The URL to send the POST request to.
            data (Optional[Union[Dict[str, Any], Any]]): JSON data to include in the request body. Defaults to None.
            headers (Optional[Dict[str, str]]): Headers to include in the request. Defaults to None.

        Returns:
            str: The response as a string.
        """
        async with aiohttp.ClientSession() as session:
            async with session.post(url, json=data, headers=headers) as resp:
                return await resp.text()
