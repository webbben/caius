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
