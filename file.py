import json
import math
import time
import traceback
from datetime import datetime, timedelta
from math import sqrt
from pickle import dumps, loads
from random import choice, randint, random
from threading import Thread
from time import sleep
import re
import os
import sys
from telegram import Update, Message
import telegram
from telegram.ext import Updater, CommandHandler
import config

DANIIL_SHARKO_ID = 84152223
FEATURE_PATTERN = re.compile(r'/set\s(.+)')
START_TEXT = """Привет! Я бот для определения особенных людей!
Список команд:
/start — команда для слабаков, которые не могут разобраться сами
/spin — определение особенного человека
/set — установка должности особенного человека
/stat — статистика по особенным человекам
/members — показать людей, которые могут быть особенными
/chat_id — показать id чата
/join — присоединить себя ко множеству людей, которые могут быть особенными."""


def err(func):
    def wp(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except Exception as e:
            print(e)
            traceback.print_exc()

    return wp


def gen_expr():
    expr = "{} {} {}".format(randint(3, 100), choice('+-*/'), randint(3, 100))
    val = eval(expr)
    return "{} = {}".format(expr, val)


def gen_inequality():
    a = randint(3, 100)
    b = randint(3, 100)
    if a > b:
        c = '>'
    elif a < b:
        c = '<'
    else:
        c = '='
    return '{} {} {}'.format(a, c, b)


def gen_sqrt():
    a = randint(3, 100)
    return "sqrt({}) = {}...".format(a, sqrt(a))


def read_info(chat_id):
    with open(os.path.join('infos', str(chat_id))) as file:
        return json.load(file)


def write_info(chat_id, info):
    with open(os.path.join('infos', str(chat_id)), 'w') as file:
        json.dump(info, file)


def start(bot, update: Update):
    update.message.reply_text(START_TEXT)


def get_chat_id(bot, update):
    update.message.reply_text(str(update.message.chat_id))


@err
def spin(bot, update):
    message = update.message
    chat_id = str(message.chat.id)
    if not os.path.exists(os.path.join('infos', chat_id)):
        bot.send_message(message.chat.id, "Нельзя выбирать особенного человека из пустого множества.")
        return
    info = read_info(chat_id)
    if not info['users']:
        bot.send_message(message.chat.id, "Нельзя выбирать особенного человека из пустого множества.")
        return

    current_time = time.time()
    feature = info['feature']
    last_spin_time = info['last_spin_time']
    if last_spin_time < current_time - 24 * 60 * 60:
        special_human = choice(list(info['users'].keys()))
        bot.send_message(chat_id, "Итак, начинаем ультрасложный подсчёт...")
        sleep(3)
        bot.send_message(chat_id, gen_expr())
        sleep(3)
        bot.send_message(chat_id, gen_inequality())
        sleep(3)
        bot.send_message(chat_id, gen_sqrt())
        sleep(3)
        bot.send_message(chat_id, "...")
        sleep(3)
        resp_message = "Сегодня {} дня — {}! Поздравляем его!".format(feature, info['users'][special_human]['name'])
        info['users'][special_human]["count"] += 1
        info['special_human'] = special_human
        last_message = bot.send_message(str(message.chat.id), resp_message)
        try:
            bot.pin_chat_message(last_message.chat_id, last_message.message_id, disable_notification=True)
        except telegram.error.BadRequest:
            pass
    else:
        resp_message = "По уже проведённому подсчёту сегодня {} дня — {}.".format(
            feature,
            info['users'][info['special_human']]['name']
        )
        bot.send_message(str(message.chat.id), resp_message)
    info['last_spin_time'] = current_time
    write_info(chat_id, info)


def join(bot, update):
    message = update.message
    chat_id = str(message.chat.id)
    user_name = message.from_user.first_name
    user_id = str(message.from_user.id)
    if os.path.exists(os.path.join('infos', chat_id)):
        info = read_info(chat_id)
        if user_id in info['users']:
            resp_message = "{} уже есть во множестве людей, которые могут быть особенными.".format(user_name)
        else:
            resp_message = "Теперь товарищ {} может быть особенным.".format(user_name)
            info['users'][user_id] = {
                'name': user_name,
                'count': 0,
            }
            write_info(chat_id, info)
    else:
        info = {
            'users': {
                user_id: {
                    'name': user_name,
                    'count': 0,
                }
            },
            'feature': 'feature',
            'last_spin_time': 0,
            'special_human': user_id,
        }
        write_info(chat_id, info)
        resp_message = "Теперь товарищ {} может быть особенным!".format(user_name)
    bot.send_message(str(message.chat.id), resp_message)


def members(bot, update):
    message = update.message
    chat_id = str(message.chat.id)
    if not os.path.exists(os.path.join('infos', chat_id)):
        resp_message = "Никто не может быть особенным, потому что никто не присоединился."
    else:
        resp_message = "Люди, которые могут быть особенными:\n{}.".format(
            '\n'.join(user['name'] for user in read_info(chat_id)['users'].values())
        )
    bot.send_message(message.chat.id, resp_message)


@err
def set_feature(bot, update):
    message = update.message
    chat_id = str(message.chat.id)
    if not os.path.exists(os.path.join('infos', chat_id)):
        bot.send_message(message.chat.id, "Нельзя ставить фичу без людей. Используйте /join.")
        return
    if message.text == '/set':
        resp_message = "Текущая особенность: {}.".format(read_info(chat_id)['feature'])
    else:
        features = FEATURE_PATTERN.findall(message.text)
        if not features:
            return
        feature = features[0]
        info = read_info(chat_id)
        info['feature'] = feature
        write_info(chat_id, info)
        resp_message = "Особенность сменена на {}.".format(feature)
    bot.send_message(str(message.chat.id), resp_message)


def get_suffix(cnt):
    count = str(cnt)
    if len(count) >= 2:
        if count[-2] == '1':
            return ''
    if count.endswith('2') or count.endswith('3') or count.endswith('4'):
        return 'a'
    else:
        return ''


def format_name(name, max_len):
    if "𝚔𝚘𝚗𝚊𝚝𝚊" in name:
        return format_name("Мишаня", max_len)
    else:
        return str.ljust(name, max_len)


def stat(bot, update):
    message = update.message
    chat_id = str(message.chat.id)
    if not os.path.exists(os.path.join('infos', chat_id)):
        bot.send_message(message.chat.id, "Нельзя проводить статистику без человеков.")
        return
    info = read_info(chat_id)
    max_nick_len = max(len(user['name']) for user in info['users'].values()) + 1
    resp_message = "Статистика по человекам с особенностью {}:\n`{}`".format(
        info['feature'],
        '\n'.join(
            '{}: {} раз{}'.format(format_name(user['name'], max_nick_len), user['count'], get_suffix(user['count'])) for
            user in info['users'].values()),
    )
    bot.send_message(str(message.chat.id), resp_message, parse_mode=telegram.ParseMode.MARKDOWN)


def top(bot, update):
    message = update.message
    chat_id = str(message.chat.id)
    if not os.path.exists(os.path.join('infos', chat_id)):
        bot.send_message(message.chat.id, "Нельзя проводить статистику без человеков.")
        return
    info = read_info(chat_id)
    max_nick_len = max(len(user['name']) for user in info['users'].values()) + 1
    resp_message = "Топ по человекам с особенностью {}:\n`{}`".format(
        info['feature'],
        '\n'.join(
            '{}: {} раз{}'.format(format_name(user['name'], max_nick_len), user['count'], get_suffix(user['count'])) for
            user in sorted(info['users'].values(), key=lambda usr: usr['count'], reverse=True)),
    )
    bot.send_message(str(message.chat.id), resp_message, parse_mode=telegram.ParseMode.MARKDOWN)


def mishagay(bot, update):
    message = update.message
    bot.send_message(str(message.chat.id), "True")


def say(bot, update: Update):
    message = update.message
    if str(message.chat.id) == str(DANIIL_SHARKO_ID):
        try:
            chat_id = int(message.text.split('|')[-1])
            bot.send_message(chat_id, message.text[4:-len(str(chat_id)) - 1])
        except ValueError as e:
            print(e)
        # bot.send_message(-1001478599943, message.text[4:])


def reset_day(bot, update: Update):
    if str(update.message.from_user.id) != str(DANIIL_SHARKO_ID):
        return
    try:
        chat_id = str(update.message.chat.id)
        info = read_info(chat_id)
        info['last_spin_time'] = 0
        write_info(chat_id, info)
        bot.send_message(chat_id, "День был сброшен")
    except Exception:
        traceback.print_exc()


def ping(bot, update):
    message = update.message
    bot.send_message(message.chat.id, 'pong')


def error(bot, update, err):
    print("ERROR", file=sys.stderr)
    print(bot, update, err, file=sys.stderr)


COMMANDS = {
    'start': start,
    'help': start,
    'chat_id': get_chat_id,
    'spin': spin,
    'stat': stat,
    'join': join,
    'say': say,
    'top': top,
    'set': set_feature,
    'members': members,
    'reset_day': reset_day,
    'mishagay': mishagay
}


def get_chats():
    return os.listdir("infos")


def f(x):
    return 2 * math.atan(x / 1440 * 6) / math.pi * 0.4


@err
def call_spin(bot):
    while True:
        for chat_id in get_chats():
            info = read_info(chat_id)
            last_spin_time = info['last_spin_time']
            time_diff_in_minutes = (time.time() - last_spin_time) / 60
            if time_diff_in_minutes < 1440:
                continue
            if random() < 1 / (24 * 30):
                user_id = choice(list(info['users'].keys()))
                print("diff:", time_diff_in_minutes)
                print("call spin with user '{}' in chat '{}'".format(info['users'][user_id]['nickname'], user_id))
                bot.send_message(
                    chat_id,
                    "Не хочешь подёргать меня за /spin, {}?".format(info['users'][user_id]['nickname'])
                )
        sleep(60)


def main():
    updater = Updater(config.TOKEN)
    dp = updater.dispatcher
    for command, handler in COMMANDS.items():
        dp.add_handler(CommandHandler(command, handler))
    updater.bot.send_message(DANIIL_SHARKO_ID, 'telegram bot has be\ten started at {}'.format(datetime.now()))
    Thread(target=call_spin, args=(updater.bot,)).start()
    dp.add_error_handler(error)
    updater.start_polling()
    updater.idle()


if __name__ == '__main__':
    main()
