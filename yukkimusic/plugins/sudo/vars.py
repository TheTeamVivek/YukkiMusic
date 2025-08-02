from strings import pick_commands

from . import mhelp

# TODO: Extract vars logic from heroku and put it here
(
    mhelp.add(
        "en",
        (
            "<b><u>Heroku:</u></b>\n\n"
            f"<b>{pick_commands('GETVAR_COMMAND')} [Var Name]</b> - Get a config var from vars\n"
            f"<b>{pick_commands('DELVAR_COMMAND')} [Var Name]</b> - Delete a var from vars\n"
            f"<b>{pick_commands('SETVAR_COMMAND')} [Var Name] [Value]</b> - Add or update a var. Separate var and its value with a space"
        ),
    )
    .add(
        "ar",
        (
            "<b><u>هيروكو:</u></b>\n\n"
            f"<b>{pick_commands('GETVAR_COMMAND')} [اسم المتغير]</b> - جلب متغير من الإعدادات\n"
            f"<b>{pick_commands('DELVAR_COMMAND')} [اسم المتغير]</b> - حذف متغير من الإعدادات\n"
            f"<b>{pick_commands('SETVAR_COMMAND')} [اسم المتغير] [القيمة]</b> - إضافة أو تحديث متغير. افصل المتغير وقيمته بمسافة"
        ),
    )
    .add(
        "as",
        (
            "<b><u>হেৰোকু:</u></b>\n\n"
            f"<b>{pick_commands('GETVAR_COMMAND')} [ভেৰিয়েবল নাম]</b> - ভেৰিয়েবলৰ পৰা কনফিগ ভেৰিয়েবল প্ৰাপ্ত কৰক\n"
            f"<b>{pick_commands('DELVAR_COMMAND')} [ভেৰিয়েবল নাম]</b> - ভেৰিয়েবলৰ পৰা ভেৰিয়েবল মচি পেলাওক\n"
            f"<b>{pick_commands('SETVAR_COMMAND')} [ভেৰিয়েবল নাম] [মান]</b> - এটা ভেৰিয়েবল যোগ কৰক বা আপডেট কৰক। ভেৰিয়েবল আৰু ইয়াৰ মান এটা খালী স্থানৰে পৃথক কৰক"
        ),
    )
    .add(
        "hi",
        (
            "<b><u>हेरोकू:</u></b>\n\n"
            f"<b>{pick_commands('GETVAR_COMMAND')} [वेरिएबल नाम]</b> - वेरिएबल्स से एक कॉन्फ़िग वेरिएबल प्राप्त करें\n"
            f"<b>{pick_commands('DELVAR_COMMAND')} [वेरिएबल नाम]</b> - वेरिएबल्स से एक वेरिएबल हटाएं\n"
            f"<b>{pick_commands('SETVAR_COMMAND')} [वेरिएबल नाम] [मान]</b> - एक वेरिएबल जोड़ें या अपडेट करें। वेरिएबल और उसके मान को स्पेस से अलग करें"
        ),
    )
    .add(
        "ku",
        (
            "<b><u>هێرۆكو:</u></b>\n\n"
            f"<b>{pick_commands('GETVAR_COMMAND')} [ناوی گۆڕاو]</b> - گرتنی گۆڕاوێک لە ڕێکخستنەکان\n"
            f"<b>{pick_commands('DELVAR_COMMAND')} [ناوی گۆڕاو]</b> - سڕینەوەی گۆڕاوێک لە ڕێکخستنەکان\n"
            f"<b>{pick_commands('SETVAR_COMMAND')} [ناوی گۆڕاو] [نرخ]</b> - گۆڕاوێک زیاد بکە یان نوێ بکەرەوە. گۆڕاوەکە و نرخی بە بۆشایی جیا بکەرەوە"
        ),
    )
    .add(
        "tr",
        (
            "<b><u>Heroku:</u></b>\n\n"
            f"<b>{pick_commands('GETVAR_COMMAND')} [Değişken Adı]</b> - Değişkenlerden bir yapılandırma değişkeni al\n"
            f"<b>{pick_commands('DELVAR_COMMAND')} [Değişken Adı]</b> - Değişkenlerden bir değişken sil\n"
            f"<b>{pick_commands('SETVAR_COMMAND')} [Değişken Adı] [Değer]</b> - Bir değişken ekle veya güncelle. Değişken ile değerini boşlukla ayır"
        ),
    )
)
