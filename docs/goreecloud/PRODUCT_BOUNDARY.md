# Product Boundary

## Purpose

GoreeCloud Video is the dedicated video-streaming application within GoreeCloud Suite. It is intended to organize, discover, stream, and manage movies, television shows, home videos, family media, and other approved video content for individual users and family profiles.

## First-class video libraries

The planned product surface includes:

- Movies
- TV Shows
- Home Videos
- Family Media
- Other Videos

## Explicit exclusions

GoreeCloud Video is not intended to become a general-purpose:

- music-streaming service
- audiobook library
- ebook reader
- podcast library
- photo-backup service

Audio remains supported when it belongs to video playback, including alternate audio tracks, commentary, descriptive audio, trailers, and similar video-associated streams.

During the initial maintained-fork stage, inherited Jellyfin code may still contain excluded product surfaces. Their presence in the inherited baseline does not make them approved GoreeCloud Video features. Removal or isolation is part of the planned transition work.

## User experience

The long-term primary client is a first-party Glaze UI experience designed for web, mobile, desktop, tablet, and television use. Jellyfin Web may remain useful as an implementation and compatibility reference during transition, but it is not the intended permanent GoreeCloud Video interface.

## Data authority

Original media files remain authoritative outside the application database. GoreeCloud Video should index and consume approved media storage through least-privilege access rather than treating its database or caches as the authoritative media copy.
