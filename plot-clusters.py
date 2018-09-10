#!/usr/bin/env python
#
# plot-clusters: plot results of entropy clustering
#
# Copyright (C) 2018 Pawel Foremski <pjf@foremski.pl>
# Licensed under GNU GPL v3
#

import sys
import matplotlib.pyplot as plt
from matplotlib import cm
import matplotlib.gridspec as gridspec
import matplotlib.ticker as plticker
import numpy as np

import argparse
p = argparse.ArgumentParser()
p.add_argument('--save')
p.add_argument('--type')
p.add_argument('-s', type=int, default=-1)
p.add_argument('-l', type=int, default=-1)
args = p.parse_args()

# read clusters
clusters = []
for line in sys.stdin:
	# strip
	line = line.strip()
	if len(line) == 0: continue

	# infer -F
	if line[0] == '2'and args.s < 0:
		try:
			j = line[:-1].index(')')
			i = line[:j].rindex('(')
			args.s = 10 + line[i:j].count(' ')
		except:
			args.s = 9

	if line[0:7] != "cluster": continue
	line = line[13:]

	p = line.index("]")
	profile = [float(x) for x in line[:p].split(" ")]
	details = [float(x) for x in line[p+2:].split(" ")]
	clusters.append({
		"P": profile,
		"%": details[0],
		"E": details[1]
	})

# infer location
if args.s < 0: args.s = 9
if args.l < 0: args.l = len(clusters[0]["P"])

# prepare the canvas
plt.figure(figsize=(8,3), dpi=100)
plt.gcf().subplots_adjust(bottom=0.18)
gs = gridspec.GridSpec(1, 4, wspace=0.09)
ax1 = plt.subplot(gs[0, 0])
ax2 = plt.subplot(gs[0, 1:])

# show colormap (hack!)
foo = ax2.imshow(np.array([np.arange(0, 1.01, 0.01)]), visible=False, aspect='auto')
cb = plt.colorbar(foo, pad=0.02, ticks=np.arange(0, 1.01, 0.2))
cb.set_label("Median entropy")
ax2.clear()

# tweak ax1
ax1.invert_xaxis()
ax1.xaxis.set_label_text(args.type + " [%]" if args.type else "Prefixes [%]")
ax1.xaxis.set_major_locator(plticker.MultipleLocator(10))
ax1.xaxis.set_minor_locator(plticker.MultipleLocator(5.0))
ax1.xaxis.grid(which='both')

ax1.set_ylim(len(clusters)+0.5, 0.5)
#ax1.yaxis.set_major_locator(plticker.MultipleLocator(base=1.0))
ax1.yaxis.set_ticks(range(1, len(clusters)+1))
ax1.yaxis.set_label_text("Cluster ID")

# tweak ax2
ax2.xaxis.set_ticks(range(args.s, args.s+args.l))
if args.l > 8:
	labels = []
	for x in range(args.l):
		if x % 2 == 0:
			labels.append("")
#			labels.append("%d" % (args.s + x))
		else:
			labels.append("%d" % (args.s + x))
#			labels.append("\n%d" % (args.s + x))
	ax2.xaxis.set_ticklabels(labels)
ax2.xaxis.set_label_text("IPv6 nybble (hex character)")

ax2.set_ylim(len(clusters)+0.5, 0.5)
ax2.yaxis.set_ticks(range(1, len(clusters)+1))
#ax2.yaxis.tick_right()
ax2.yaxis.set_ticklabels([])

# draw entropy profiles
#ax2.imshow(np.array([x["P"] for x in clusters]), interpolation='nearest', aspect='auto')
foo = None
for y,cluster in enumerate(clusters):
	for x,ent in enumerate(cluster["P"]):
		ax2.barh(y+1, 1.0, 0.8, args.s+x-0.5, align='center', color=cm.jet(ent))

	ax1.barh(y+1, cluster["%"], 0.4, align='center', color='red')

# hide zero on ax1.X
#plt.setp(ax1.get_xticklabels()[0], visible=False)

# show it
#plt.tight_layout()
if args.save:
	plt.savefig(args.save)
	print "Saved to %s" % (args.save)
else:
	plt.show()
