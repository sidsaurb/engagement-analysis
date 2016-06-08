#!/usr/bin/env python3

import time
import subprocess
import platform
import argparse
from termcolor import colored
import utils
import requests

PING_DELAY = 10
LOCAL_HOST = 'localhost'
LOCAL_HOST_TIMEOUT = 1
REMOTE_HOST_TIMEOUT = 2

PING_COUNT_FLAG = '-n' if platform.system().lower() == 'windows' else '-c'

STATE_OK = 0
STATE_LOCAL_ERROR = 1
STATE_CONNECTION_ERROR = 2
STATE_REMOTE_ERROR = 3
STATE_UNKNOWN = 4

SEND_EMAIL = False
SEND_SMS = False

EMAIL = 'g10pahal@gmail.com'
PHONE_NO = '+918953454203'


def shell_ping(host, timeout):
    exit_code = subprocess.call(
            ['ping', PING_COUNT_FLAG, '1', '-W', str(timeout), host],
            stdout=open('/dev/null', 'w'),
            stderr=open('/dev/null', 'w')
    )

    return exit_code == 0


def ping_basic(host):
    local_ping = shell_ping(LOCAL_HOST, LOCAL_HOST_TIMEOUT)
    if not local_ping:
        return STATE_LOCAL_ERROR

    remote_ping = shell_ping(host, REMOTE_HOST_TIMEOUT)
    if not remote_ping:
        return STATE_REMOTE_ERROR

    return STATE_OK


def ping(host):
    local_ping = shell_ping(LOCAL_HOST, LOCAL_HOST_TIMEOUT)
    if not local_ping:
        return STATE_LOCAL_ERROR

    try:
        r = requests.get(host + '/ping')
        if r.status_code != 200 or r.text != 'success':
            return STATE_REMOTE_ERROR
    except:
        return STATE_REMOTE_ERROR

    return STATE_OK


def internet_connected():
    return ping_basic('google.com') == STATE_OK


def ping_multiple(hosts, remote=True):
    result = []
    for host in hosts:
        ping_basic_result = ping(host)
        if remote and ping_basic_result == STATE_REMOTE_ERROR and not internet_connected():
            result.append(STATE_CONNECTION_ERROR)

        result.append(ping_basic_result)

    return result


def sprint_color(before, heading, after, color):
    return colored(before, color) + colored(heading, color, attrs=['bold']) + colored(after, color)


def print_color(before, heading, after, color, **kwargs):
    return print(sprint_color(before, heading, after, color), **kwargs)


def print_state_line(title, content, color):
    print_color('', title + ': ', content, color, end='')


def handle_status_change(host, title, content, initial):
    if not initial:
        if SEND_EMAIL:
            utils.EmailThread(EMAIL, 'Ping service status for host {}'.format(host), ' => ' + title + ': ' + content).start()
        if SEND_SMS:
            utils.SMSThread(PHONE_NO, 'Ping service status for host {}'.format(host) + ' => ' + title + ': ' + content).start()


def print_state(state, next_state, hosts, initial):
    index = 0
    print('[', end='')

    for host in hosts:
        if index != 0:
            print(', ', end='')

        if next_state[index] == STATE_OK:
            if state[index] != next_state[index]:
                handle_status_change(host, 'Machine Status', 'running', initial)
            print_state_line('Machine Status', 'running', 'green')
        elif next_state[index] == STATE_LOCAL_ERROR:
            if state[index] != next_state[index]:
                handle_status_change(host, 'Local Network', 'error in local network interface', initial)
            print_state_line('Local Network', 'error in local network interface', 'yellow')
        elif next_state[index] == STATE_CONNECTION_ERROR:
            if state[index] != next_state[index]:
                handle_status_change(host, 'Internet Connection', 'error in internet connection', initial)
            print_state_line('Internet Connection', 'error in internet connection', 'yellow')
        elif next_state[index] == STATE_REMOTE_ERROR:
            if state[index] != next_state[index]:
                handle_status_change(host, 'Machine Status', 'not running', initial)
            print_state_line('Machine Status', 'not running', 'red')
        else:
            if state[index] != next_state[index]:
                handle_status_change(host, 'Ping Service', 'error encountered - unknown state', initial)
            print_state_line('Ping Service', 'error encountered - unknown state', 'yellow')

        index += 1

    print(']' + ' ' * 20 * len(hosts), end='\r')


def ping_service(hosts, remote=True):
    print('Starting ping service for hosts {} ...'.format(hosts))
    print('')

    count = len(hosts)

    state = [STATE_UNKNOWN] * count
    next_time = time.time()

    initial = True

    try:
        while True:
            next_state = ping_multiple(hosts, remote)
            if state != next_state:
                print_state(state, next_state, hosts, initial)
                state = next_state

            next_time += PING_DELAY
            time.sleep(next_time - time.time())

            initial = False
    except:
        print('\n\nShutting down the ping service ...')


if __name__ == '__main__':
    default_hosts = ['http://localhost:8080', 'http://localhost:8081', 'http://localhost:8083']

    parser = argparse.ArgumentParser(prog='ping_monitor.py',
                                     description='Start a ping service to monitor health of several servers.')
    parser.add_argument('--hosts', nargs='+', default=default_hosts, metavar='HOST',
                        help='hosts on which the ping service is run')
    parser.add_argument('--email', nargs='?', const=EMAIL, default=None, metavar='EMAIL',
                        help='email address to notify')
    parser.add_argument('--sms', nargs='?', const=PHONE_NO, default=None, metavar='PHONE_NO',
                        help='phone number to notify')
    args = parser.parse_args()

    if args.email is None:
        SEND_EMAIL = False
    else:
        SEND_EMAIL = True
        EMAIL = args.email

    if args.sms is None:
        SEND_SMS = False
    else:
        SEND_SMS = True
        PHONE_NO = args.sms

    ping_service(args.hosts)
