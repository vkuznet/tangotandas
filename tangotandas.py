#!/usr/bin/env python
#-*- coding: utf-8 -*-
#pylint: disable=
"""
File       : itunes.py
Author     : Valentin Kuznetsov <vkuznet AT gmail dot com>
Description: 

NOTE: to make this script work track names/artists/genres should not contains commas
"""

# system modules
from __future__ import print_function
import os
import sys
import time
import argparse
import datetime
import dateutil.relativedelta
import subprocess
from subprocess import PIPE

START_TIME = time.time()

# colors: http://www.tldp.org/HOWTO/Bash-Prompt-HOWTO/x329.html
#         http://ascii-table.com/ansi-escape-sequences.php
BLACK, RED, GREEN, BROWN, BLUE, PURPLE, CYAN, LIGHT_GRAY = \
        ['0;%dm'%(30+i) for i in range(8)]
DARK_GRAY, LIGHT_RED, LIGHT_GREEN, YELLOW, LIGHT_BLUE, LIGHT_PURPLE, LIGHT_CYAN, WHITE =\
        ['1;%dm'%(30+i) for i in range(8)]
BOLD = '\x1b[1m'
BKG_WHITE = '\x1b[47m'

MUSIC_NOTE = u'\u266B'
MUSIC_COUPLE = u'\u8E40'
MUSIC_BREAK2 = u'\u1F377'
MUSIC_BREAK = u'\u2615'

class OptionParser():
    def __init__(self):
        "User based option parser"
        self.parser = argparse.ArgumentParser(prog='PROG')
        self.parser.add_argument("--interval", action="store",
            dest="interval", default=1, help="Refresh interval")
        self.parser.add_argument("--time-offset", action="store",
            dest="time_offset", default=0, help="Timestamp of start-time")

def color(text, col):
    "Return colored text"
    seq = BOLD + "\x1b[%s" % col + text + "\x1b[0m"
    return seq

def exe(cmd):
    "Execute command"
    res = subprocess.Popen(cmd, shell=True, stdin=PIPE, stdout=PIPE, stderr=PIPE, close_fds=True)
    return res.stdout.read()

def tracks():
    "Fetch all tracks from current playlist"
    cmd = 'tell application "iTunes" to get name of every track in current playlist'
    cmd = "osascript -e '%s'" % cmd
    return exe(cmd).split(', ')

def artists():
    "Fetch all artists from current playlist"
    cmd = 'tell application "iTunes" to get artist of every track in current playlist'
    cmd = "osascript -e '%s'" % cmd
    return exe(cmd).split(', ')

def genres():
    "Fetch all genres from current playlist"
    cmd = 'tell application "iTunes" to get genre of every track in current playlist'
    cmd = "osascript -e '%s'" % cmd
    return exe(cmd).split(', ')

def track():
    "Fetch current track"
    cmd = """tell application "iTunes"
tell current track to artist & tab & name & tab & genre
end tell
"""
    cmd = "osascript -e '%s'" % cmd
    trk = exe(cmd)
    artist, track, genre = trk.split('\t')
    trk = track.strip().replace('\n', '')
    track = unicode(trk, 'utf-8')
    return dict(artist=artist, track=track, genre=genre.replace('\n', ''))

def print_song(song, col=None, pad=""):
    "Human readable output for given song"
    if song['genre'] == 'Cortina':
        track = song['track'].replace('z_', '').encode('utf-8')
        msg = "%sCortina: %s, %s" % (pad, song['artist'], track)
    else:
        track = song['track'].encode('utf-8')
        msg = "%s%s, %s" % (pad, song['artist'], track)
    if  col:
        print(color(msg, col))
    else:
        print(msg)

def print_tanda(tanda, trk=None, prefix="CURRENT TANDA"):
    "Human readable format for tanda"
    title = None
    for song in tanda:
        genre = song['genre'].lower()
        pad = '  '
        if  not title:
            title = "%s: %s" % (prefix, genre.capitalize())
            print(title)
        if song == trk:
            pad = MUSIC_NOTE.encode('utf-8') + ' '
            if  genre == 'tango':
                col = RED
            elif genre == 'vals':
                col = GREEN
            elif genre == 'milonga' or genre == 'tango foxtrot':
                col = PURPLE
            else:
                col = LIGHT_CYAN
                pad = MUSIC_BREAK.encode('utf-8') + ' '
        else:
            col = BLACK
#        if  not title:
#            title = "%s: %s" % (prefix, genre.capitalize())
#            print(color(title, col))
        print_song(song, col=col, pad=pad)

def playlist(time_offset=0):
    "Fetch playlist from itunes"
    itracks = tracks()
    iartists = artists()
    igenres = genres()
    itrack = track()
    ptracks = []
    if  len(itracks) != len(iartists) or \
        len(itracks) != len(igenres) or \
        len(iartists) != len(igenres):
        print("Please check your playlist, # tracks %s, # artists %s, # genres %s" \
                % (len(itracks), len(iartists), len(igenres)))
        print("Common error if some field contains commas")
        sys.exit(1)
    for idx in range(len(itracks)):
        row = {}
        trk = itracks[idx].strip().replace('\n', '')
        row['track'] = unicode(trk, 'utf-8')
        row['artist'] = iartists[idx].strip().replace('\n', '')
        row['genre'] = igenres[idx].strip().replace('\n','')
        ptracks.append(row)
    tanda = []
    old_tanda = []
    exit = False
    for row in ptracks:
        if  row['track'] == itrack['track']:
            exit = True
        if  row['genre'] == 'Cortina' and exit:
            tanda.append(row)
            break
        if  row['genre'] == 'Cortina':
            old_tanda = tanda
            tanda = []
            continue
        tanda.append(row)
    if  exit:
        if  old_tanda != tanda:
            clear()
            print_tanda(tanda, itrack)
            # find next tanda
            next_tanda = []
            for idx in range(len(ptracks)):
                row = ptracks[idx]
                if  row == tanda[-1]:
                    for jdx in range(len(ptracks)):
                        if  jdx <= idx:
                            continue
                        row = ptracks[jdx]
                        next_tanda.append(row)
                        if  row['genre'] == 'Cortina':
                            break
            print_tanda(next_tanda, None, '\nNEXT TANDA')
            print(time_reminder(time_offset))

def time_reminder(time_offset=0):
    dt1 = datetime.datetime.fromtimestamp(START_TIME-time_offset)
    dt2 = datetime.datetime.fromtimestamp(time.time())
    tdiff = dateutil.relativedelta.relativedelta (dt2, dt1)
    lst = "%d hours, %d minutes and %d seconds" \
            % (tdiff.hours, tdiff.minutes, tdiff.seconds)
    msg = '\nDJ Valentin: %s' % lst
    return color(msg, LIGHT_PURPLE)

def clear():
    "Clear terminal screen"
#    cmd = "tput reset"
#    exe(cmd)
    print(chr(27) + "[2J")

def main():
    "Main function"
    optmgr  = OptionParser()
    opts = optmgr.parser.parse_args()
    while True:
        playlist(int(opts.time_offset))
        time.sleep(int(opts.interval))

if __name__ == '__main__':
    main()
