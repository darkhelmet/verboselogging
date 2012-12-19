--- 
id: 474
author: Daniel Huckstep
title: Instant NoSQL Cluster With Chef, Cassandra, And Your Favorite Cloud Hosting Provider
category: software
description: Crank up a Cassandra NoSQL cluster with the magic of Chef and cloud providers.
published: true
publishedon: 05 Nov 2010 08:00 MDT
slugs: 
- instant-nosql-cluster-with-chef-cassandra-and-your-favorite-cloud-hosting-provider
tags: 
- chef
- cassandra
- nosql
- rackspace
- amazon
images: 
  dr_nick: 
    small: http://cdn.verboselogging.com/transloadit/small/ba/ee1338a73e50dae3cb628fd414d180/dr-nick.gif
    medium: http://cdn.verboselogging.com/transloadit/medium/ef/943cb0c0aa1f0e47175ed7b2e5eea5/dr-nick.gif
    original: http://cdn.verboselogging.com/transloadit/original/b0/6426354b2b4f5b1b0042a3e1e1ac93/dr-nick.gif
    large: http://cdn.verboselogging.com/transloadit/large/8f/f861556e7576616711011b4b343bde/dr-nick.gif
---
<p><figure><img src="http://cdn.verboselogging.com/transloadit/medium/ef/943cb0c0aa1f0e47175ed7b2e5eea5/dr-nick.gif" class="fright bbottom bleft round medium" alt="" /></figure></p>
<h2>Hi everybody!</h2>
<p>Okay, so don&#8217;t worry, I&#8217;m not <a href="http://en.wikipedia.org/wiki/Dr._Nick_Riviera">Dr. Nick Riviera</a>, I&#8217;m not going to take your liver out. Well, not unless you need to sell it!</p>
<p>I am going to tell how to get your NoSQL on with a little bit of Cassandra, a little bit of Chef, and a little bit of sensual.. NO NO NO! Nevermind, none of that.</p>
<h2>Seriously</h2>
<p>No really, we&#8217;re going to get some <a href="http://www.xtranormal.com/watch/6995033/">/dev/null</a> web scale up <a href="http://www.explosm.net/db/files/Comics/Rob/upinthisbitch.png">in this bitch</a></p>
<p>But not with MongoDB. This setup is more suited to a Dynamo style system, and not a master-master, or replica system like CouchDB or MongoDB.</p>
<h2>No really, seriously let&#8217;s do this</h2>
<p>Okay enough shtick. Let&#8217;s do this. Before we get too far, I should tell I&#8217;m not going to teach you how to use Chef. Or how to configure Cassandra. Those are other balls of wax. I&#8217;ll just show you cool stuff you can do with a really sweet feature of Chef to crank up your cluster.</p>
<h2>What&#8217;s so cool about Chef?</h2>
<p>Chef is pretty cool. It&#8217;s along the same lines as <a href="http://www.puppetlabs.com/">puppet</a> if you&#8217;ve used that. It has a central server which keeps track of hosts (nodes in Chef-speak), and the really cool feature I was talking about is the ability to search your nodes when you are setting one up. You can do some stuff like this:</p>
<pre>search(:node, 'name:db*')</pre>
<p>in a recipe to get all the nodes whose name starts with &#8220;db&#8221;. Awesome! You could set up <code>iptables</code> to only allow connections from the hosts in your network. You could&#8230;um&#8230;do some pretty cool stuff. You really can. We&#8217;re going to use this feature to setup our cluster.</p>
<h2>Searching, <span class="caps">LIKE</span> A <span class="caps">BOSS</span>!</h2>
<p>With Cassandra, you setup a single node, and it has <em>seeds</em>. Well okay, it really only needs one to get started. The seeds are just nodes it <em>gossips</em> with to figure out where everything is. Then it can continue on its merry way, migrating data around, scaling your app by <a href="http://jamesgolick.com/2010/10/27/we-are-experiencing-too-much-load-lets-add-a-new-server..html">adding another server</a>.</p>
<p>So. Where does Chef search come in?</p>
<p>When we crank up another Cassandra node, we can search for and find all the other nodes in the Cassandra cluster, and set those as the seeds so it can gossip like the stereotypical office secretary. It not super exciting since it will figure out where all the nodes are given just one, but it&#8217;s good stuff anyway. Cassandra also has a very open security model (was designed for situations where you control the <span class="caps">LAN</span>, <span class="caps">AFAIK</span>), so that thing I talked about before? Setting up iptables to only accept connections from a certain set of nodes? Pretty useful now, isn&#8217;t it!</p>
<h2>Show me some code already!</h2>
<p>Alright, but only a peak!</p>
<script src="https://gist.github.com/662132.js?file=default.rb"></script><p>This is the relevant chunk of the <code>default.rb</code> from the Chef recipe. We search for the nodes that have similar names (I was going for the cassandra1, cassandra2 kind of setup), grab their private IP address (you can adapt this for Amazon or Rackspace, or both), and throw these in the config file for the seeds value.</p>
<p>That search is <strong>everything</strong>. I&#8217;m just using it to setup the seeds, but you can use it for iptables, for setting up replication between CouchDB or MySQL servers, or setting up an nginx load balancer. You just search for the relevant nodes you need, and away you go.</p>
<p>Maybe it&#8217;s just the wine, but I&#8217;m excited about that search. It&#8217;s not like it&#8217;s new, it didn&#8217;t just come out in a new version of Chef, but it&#8217;s damn exciting. This isn&#8217;t your dad&#8217;s <a href="http://en.wikipedia.org/wiki/DevOps">devops</a>, it&#8217;s automation cranked up to 11.</p>