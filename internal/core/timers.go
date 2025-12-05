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
package core

import "time"

type scheduledTimers struct {
	scheduledUnmuteTimer *time.Timer
	scheduledResumeTimer *time.Timer
	scheduledSpeedTimer  *time.Timer
	scheduledUnmuteUntil time.Time
	scheduledResumeUntil time.Time
	scheduledSpeedUntil  time.Time
}

func (st *scheduledTimers) RemainingUnmuteDuration() time.Duration {
	if st == nil || st.scheduledUnmuteUntil.IsZero() {
		return 0
	}
	return time.Until(st.scheduledUnmuteUntil)
}

func (st *scheduledTimers) RemainingResumeDuration() time.Duration {
	if st == nil || st.scheduledResumeUntil.IsZero() {
		return 0
	}
	return time.Until(st.scheduledResumeUntil)
}

func (st *scheduledTimers) RemainingSpeedDuration() time.Duration {
	if st == nil || st.scheduledSpeedUntil.IsZero() {
		return 0
	}
	return time.Until(st.scheduledSpeedUntil)
}

func (st *scheduledTimers) cancelScheduledUnmute() {
	if st != nil && st.scheduledUnmuteTimer != nil {
		st.scheduledUnmuteTimer.Stop()
		st.scheduledUnmuteTimer = nil
		st.scheduledUnmuteUntil = time.Time{}
	}
}

func (st *scheduledTimers) cancelScheduledResume() {
	if st != nil && st.scheduledResumeTimer != nil {
		st.scheduledResumeTimer.Stop()
		st.scheduledResumeTimer = nil
		st.scheduledResumeUntil = time.Time{}
	}
}

func (st *scheduledTimers) cancelScheduledSpeed() {
	if st != nil && st.scheduledSpeedTimer != nil {
		st.scheduledSpeedTimer.Stop()
		st.scheduledSpeedTimer = nil
		st.scheduledSpeedUntil = time.Time{}
	}
}
