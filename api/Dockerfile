FROM nvidia/cuda:12.9.0-cudnn-devel-ubuntu22.04
LABEL authors="str20tbl"

ENV CUDA_VISIBLE_DEVICES="0"
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update
RUN apt-get install --no-install-recommends --no-install-suggests -y \
build-essential curl espeak-ng libsndfile1-dev git unzip \
python3 python3-dev python3-venv python3-pip sox ffmpeg

RUN python3 --version

ENV NUMBA_CACHE_DIR=.numba_cache

WORKDIR /workspace

COPY . /workspace

ENV VIRTUAL_ENV=/opt/venv
RUN python3 -m venv $VIRTUAL_ENV
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

RUN pip3 install --upgrade pip
RUN pip3 install -r requirements.txt
RUN pip3 install torch --extra-index-url https://download.pytorch.org/whl/cu124