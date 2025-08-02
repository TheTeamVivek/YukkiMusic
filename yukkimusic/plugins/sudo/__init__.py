#
# Copyright (C) 2024-2025 by TheTeamVivek@Github, < https://github.com/TheTeamVivek >.
#
# This file is part of < https://github.com/TheTeamVivek/YukkiMusic > project,
# and is released under the MIT License.
# Please see < https://github.com/TheTeamVivek/YukkiMusic/blob/master/LICENSE >
#
# All rights reserved.
#

from yukkimusic.core.help import ModuleHelp

mhelp = ModuleHelp("Sudoers")

(
    mhelp.name("en", "Sudoers")
    .name("ar", "سودو")
    .name("as", "সুডো")
    .name("hi", "सुडो")
    .name("ku", "سودۆ")
    .name("tr", "Sudo")
)


ohelp = ModuleHelp("Owner")
(
    ohelp.name("en", "Owner")
    .name("ar", "المالك")
    .name("as", "মালিক")
    .name("hi", "मालिक")
    .name("ku", "خاوەن")
    .name("tr", "Sahip")
    .add(
        "en",
        "<b><u>Owner Privileges:</u></b>\n\nOwners can use all <b>Sudoers</b> and <b>Dev</b> commands.",
        priority=100,
    )
    .add(
        "ar",
        "<b><u>صلاحيات المالك:</u></b>\n\nيمكن للمالكين استخدام جميع أوامر <b>المدراء</b> و<b>التطوير</b>.",
        priority=100,
    )
    .add(
        "as",
        "<b><u>মালিকৰ বিশেষাধিকাৰ:</u></b>\n\nমালিকে সকলো <b>সুডোৰ্ছ</b> আৰু <b>ডেভ</b> কমান্ড ব্যৱহাৰ কৰিব পাৰে।",
        priority=100,
    )
    .add(
        "hi",
        "<b><u>मालिक की विशेष अनुमति:</u></b>\n\nमालिक सभी <b>सूडोअर्स</b> और <b>डेव</b> कमांड का उपयोग कर सकते हैं।",
        priority=100,
    )
    .add(
        "ku",
        "<b><u>مۆڵەتی خاوەن:</u></b>\n\nخاوەن دەتوانن هەموو فەرمانەکانی <b>سودوەکان</b> و <b>گەشەپێدەر</b> بەکاربێنن.",
        priority=100,
    )
    .add(
        "tr",
        "<b><u>Sahip Yetkileri:</u></b>\n\nSahipler tüm <b>Sudoers</b> ve <b>Dev</b> komutlarını kullanabilir.",
        priority=100,
    )
)
