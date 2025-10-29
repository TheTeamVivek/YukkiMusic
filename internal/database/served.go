/*
 * This file is part of YukkiMusic.
 *
 * YukkiMusic — A Telegram bot that streams music into group voice chats with seamless playback and control.
 * Copyright (C) 2025 TheTeamVivek
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */
package database

func GetServed(user ...bool) ([]int64, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	state, err := getBotState(ctx)
	if err != nil {
		return nil, err
	}

	if len(user) > 0 && user[0] {
		return state.Served.Users, nil
	}
	return state.Served.Chats, nil
}

func IsServed(id int64, user ...bool) (bool, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	state, err := getBotState(ctx)
	if err != nil {
		return false, err
	}

	target := state.Served.Chats
	if len(user) > 0 && user[0] {
		target = state.Served.Users
	}

	for _, v := range target {
		if v == id {
			return true, nil
		}
	}
	return false, nil
}

func AddServed(id int64, user ...bool) error {
	exists, err := IsServed(id, user...)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	ctx, cancel := mongoCtx()
	defer cancel()

	state, err := getBotState(ctx)
	if err != nil {
		return err
	}

	target := &state.Served.Chats
	if len(user) > 0 && user[0] {
		target = &state.Served.Users
	}

	*target = append(*target, id)
	return updateBotState(ctx, state)
}

func DeleteServed(id int64, user ...bool) error {
	exists, err := IsServed(id, user...)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	ctx, cancel := mongoCtx()
	defer cancel()

	state, err := getBotState(ctx)
	if err != nil {
		return err
	}

	target := &state.Served.Chats
	if len(user) > 0 && user[0] {
		target = &state.Served.Users
	}

	newSlice := make([]int64, 0, len(*target))
	for _, v := range *target {
		if v != id {
			newSlice = append(newSlice, v)
		}
	}
	*target = newSlice
	return updateBotState(ctx, state)
}
