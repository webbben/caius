# 6/16/2025

Today I added some metrics tools for measuring performance of the different LLMs. One of the bottlenecks I've run into is performance, since it can take quite a long time to analyze large directories that have 20, 50, 80 etc files. So, I decided to build out a system that will automatically track the average execution times of each function that uses LLMs, and every LLM call itself.

I've realized now that this is going to be very important. After running some basic tests on the various models I have been experimenting with up until now, I have some pretty interesting results. Here's the data output from test runs on a ~200 line file of mock javascript code:

```
Llama3
call count: 10
Ave/Min/Max call time: 2062 ms / 1503 ms / 7056 ms

DeepSeek
call count: 10
Ave/Min/Max call time: 3463 ms / 2357 ms / 13230 ms

DeepSeek14b
call count: 10
Ave/Min/Max call time: 7882 ms / 5548 ms / 27539 ms

CodeLlama
call count: 10
Ave/Min/Max call time: 5844 ms / 4628 ms / 16686 ms

CodeLlama13b
call count: 10
Ave/Min/Max call time: 8756 ms / 6415 ms / 28898 ms
```

From these results, it's clear that `Llama3` is the fastest (the 7 second case seemed to be a bit of an outlier that happened on first execution).
Interestingly, `DeepSeek` performed second best. I thought DeepSeek would be less performant since it usually spends some time thinking to itself.

`CodeLlama` was actually perhaps the worst performing model; the smaller model took 5-6 seconds to analyze the file, and the 13b model took almost 9 seconds on average. I recently made the switch to using CodeLlama, so this was very useful to see. If it turns out that other models like Llama3 or Deepseek perform just as well as CodeLlama for the analysis tasks (like making a description for a file) then I will definitely be switching away from CodeLlama. Perhaps CodeLlama is more well suited for writing code rather than reading and describing.

> I also remember reading the documentation for CodeLlama, and it seemed to have some specific syntax for different prompt types. Maybe I need to look into that.

I recently found out there is a `DeepSeek-coder` model out now - I saw a random article about it when searching for "deepseek vs codellama". So I think next I'll wanna get that model into my test suite to find out how its performance compares to the other models.

Also, this test only considers response time. I'm going to need to come up with a way to also simultaneously test the accuracy, so that I can see if there are tradeoffs there too. Maybe just log the answer each model gives, and compare them myself?

Next things to do:

- test `DeepSeek-coder`
- switch to using faster models like Llama3 or DeepSeek (stop using CodeLlama if accuracy doesn't deteriorate)

# 5/30/2025

At the point of writing, we have some basic functionality implemented for describing files and entire projects. For describing individual files, it just passes the source code directly into an LLM to get a description. For describing an entire project, it runs the individual file analysis on each file, and records the description of each file (in 1 or 2 sentences, roughly). Then, it feeds all the file paths and their descriptions into an LLM to assess what the entire project does overall. It actually works quite well.

Things I've learned so far:

- Ollama supports specific JSON formatting for the output, which is immensely useful for coding AI tools. So, if you want specific information, you can define the schema, and the AI will return the data in that specific schema, from which your code can parse the data.
  - One challenge that still persists though, is ensuring the LLM understands which field in the JSON is for which piece of data. It's a good idea to find a way to explain this in the prompt. But, overall it seems to always give you the correctly formatted output.
- It generally seems good to break down LLM problems into sub problems. What I mean by this is, rather than asking the LLM to tell you A, B, and C about a given file, try to figure out sub-tasks and process them one at a time, providing the result in the next prompt. This, at least, seems to improve accuracy, since it seems like low-powered LLMs can get lost or forget to do things sometimes.
  - Caveat: with each "sub-problem" you break the problem into, you are multiplying the processing time. So, you want to strike a balance. Only do this for tasks where accuracy is very important.
  - For "high-cost" processing tasks (perhaps where more than one sub-task is involved), it's a good idea to cache the data somewhere. I think using more storage space is always better than having to recalculate things repeatedly, due to how "slow" locally run LLMs are. Of course, if you have a beast computer with high spec GPUs, this becomes less important I suppose.
- When possible, avoid using LLMs. They are just very costly for processing tasks. For example, rather than asking the LLM to tell you what programming language a snippet of code is, write some code that does that automatically, e.g. by looking at file extensions. Never be lazy and opt to have an LLM do the task if you think it's feasible to do it programmatically and with high accuracy. Even a small LLM prompt on my macbook pro will usually take at least a second or two, which is an eternity compared to the <= 1ms that regular code based logic would take.

Challenges so far:

LLM prompting can be kind of fickle, and feels a bit "unscientific" at times. For example, adjusting just a few words can have a significant impact, but there isn't any clear guide on what makes the best prompt. Formatting, wording, etc. And I'm guessing it varies depending on the model. Anyway, this is very different from programming, where things work more strictly and have very predictable outcomes. Another thing is, when you adjust a prompt, sometimes it will work better for certain test data, but then get slightly worse for others. I think this typically happens when you have multiple "sub-problems" baked into a single prompt - so perhaps it indicates I'm on the right track for looking to split up complicated analysis problems when possible. For example, instead of asking the AI to determine the programming language of a snippet of code AND what it does, perhaps it's good to first determine the programming language, and then ask the the LLM (given that information) what the code snippet does.

Next tasks:

Next, I think I want to work on a Q&A chat feature. Since it seems the analysis of a project works well, then it would be cool to see if we can get this AI agent to accurately answer questions about the project. Perhaps this will lead to me identifying some weaknesses with how the analysis features currently work, too.
