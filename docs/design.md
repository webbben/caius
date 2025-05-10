# Design: AI Agent

This project will serve as an experiment in creating an "AI Agent". This is mainly for learning and for-fun purposes, but who knows, maybe someday this will turn into a useful tool that I can use in my work or daily life someday.

## Motivation

Well, "AI Agents" seem to be the latest buzz in the AI world. Everyone is excited to either off-load all their work to an AI agent, or replace their human workforce with them. I'm half joking - but I guess we will see how it all shapes up. But one thing that does seem likely: "AI Agents" will continue to become a more integrated part of our work lives, especially as software engineers, but perhaps in all aspects of our work culture across industries.

So, I have an interest to learn more about this and see what it's all about. I also can envision this being a fun and possibly useful side project.

## V 1.0: Overview and Features

I guess we will call this the design for V 1.0 of our agent. Let's describe overall what it aims to accomplish, and list out the major features we will make a foray into.

V 1.0 of the agent is solely designed for use as what is essentially a CLI tool that uses AI to accomplish "smart" things. "Smart" mainly means being able to process lots of text, code, etc and do some basic reasoning and do something special. In V 1.0, it won't be an autonomous agent, and will instead just start work when a command is entered in the CLI by the user. This CLI tool is solely being designed (as of V 1.0) for **programming/software engineering tasks**. So, working in some kind of project workspace, the agent will be prompted to take some kind of action, and when its done, it will stop running.

For example, lets imagine we have a command called `new-feature` where we can prompt the agent to create a new feature. Maybe it looks something like this:

```
myAgent new-feature --name "login page" "create a login page that plugs into our existing auth API"
```

I think this specific feature idea will be far too advanced for V1 though. We will start with mostly just analysis tasks.

> In a later version (if this project continues that far), perhaps instead of needing to do all of the commands via command line, we could have an AI agent that is continuously running and taking commands in a different form, such as a chat. But that will be a much later development. By narrowing the interactivity to just CLI, we can instead focus on functionality and making the AI good at actually doing things, rather than focusing on UI/UX.

### About AI Usage

We will use Ollama, the open source locally hosted LLM engine that lets you use lots of different models that are fine tuned for various purposes. Perhaps we will use different models for different purposes, but this will take a lot of trial and error, research and practice, etc. For now, the most appealing and highly performing models overall seem to be Llama3 and DeepSeek.

### Feature: Analyze project

A fundamental feature that will be crucial for any success will be for the agent to be able to analyze and accurately understand the project it is working on. Ideally, the agent can figure everything out on its own, by looking at the directory structure, file types and code within files, and fully understand what is happening in the project overall.

From the root of the project, a user can execute `myAgent analyze .` (`.` being a path) which will prompt the AI agent to start building a directory tree map, and noting the files and their names, extensions, etc in each directory. Once it has this data, it can go deeper and start to learn about each individual file.

This process will probably be quite slow, especially if want the agent to look at each file and create summaries for their purposes, etc. So, perhaps we could make a `--fast` option which doesn't actually analyze each file in depth, and just makes assumptions based on a smaller amount of data.

The result of this will probably be some kind of document that explains the system, which the agent can use afterwards as reference while doing other tasks.
Either this, or maybe we can make a database that stores info about the project. Not sure yet.

### Feature: Code Review

Another useful feature will be having an AI agent that can essentially do code review for the user. I'm not entirely sure, but I think this could be done a few ways:

- code review of a single file; point the agent to a specific file, and tell it to review the contents of the file. perhaps even telling it something specific to review for. it will create a list of issues or possible enhancements, etc.
- code review of a PR; using the github cli, it's quite easy to show the AI a pull request, diffs of files in the PR, etc. This also makes the scope of changes to review a little more straightforward.
- code review of a specific branch, diffing with another branch; for example, if you have a branch called develop, and a feature branch that is branched off of develop, you could point the agent to review the changes that exist in the feature branch. This would be very similar to essentially reviewing a PR, but I'm not sure if the github cli would facilitate this in any way, or if we need to make our own tools for diffing the branches and all that.

This will be a very important feature for an AI agent that specializes in software projects though. It's probably one of the big use cases for an AI agent, besides writing actual code for you.

### Feature: Q&A

This feature would enable the AI agent to answer user questions about the project. First, we would need the "analyze project" feature working well, and then based on the information the agent gathers there, it would have an interactive chat with the user, answering whatever questions you pose it. Helpful for if you are entering a new project and want to figure something out, or if you need to implement something new and aren't really sure the best way to go about it.

## (Hypothetical) V 2.0:

At this point, the AI agent will have a good fundamental skillset in the following things:

- understanding a project, its architecture, technologies it is using, etc
- answering questions about a project, ideally very accurately and not making things up
- performing code review that provides some value to the user; conservative is better, as we don't want lots of excessive, unimportant feedback.

Once these things have been confirmed to be working well, and users can feel usefulness from it, it's time for V 2.0, where the AI agent starts to come to life more.
This is where the following features could start to be implemented:

- ability to implement basic features, or at least start out basic features for the user.
  - these new features would be created and put in a new branch. the user will probably expect to go in and test it themselves, confirm it's working, fix issues, etc.
  - but will save the user a lot of time hopefully.
- ability to fix bugs described in github issues
  - the user could write up issues, and the agent will read those issues and propose a fix strategy
  - if the user agrees, then the agent will move forward with implementing the fix in a new fix branch.
