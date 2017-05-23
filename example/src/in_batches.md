---
title: Rails in_batches API
date: 2015-09-20
---

At Sariina, we have a daily update job that affects millions of rows.
On this table, we also have an `ON UPDATE` trigger.
The update takes a long time and a lot of disk IO.
It makes the database virtually useless until the job is done.

I thought it would be a good idea to try and do the update in smaller batches.
I implemented a few solutions and the best one is now [merged into Rails core](https://github.com/rails/rails/pull/20933).
DHH suggested we make a new API for this.
The API looks good in my opinion.
Here are a few examples:


    Person.in_batches.each_record(&amp;:party_all_night!)
    Person.in_batches.update_all(awesome: true)
    Person.in_batches(of: 2000).delete_all
    Person.in_batches.map do |relation|
      relation.delete_all
      sleep 10 # Throttles delete queries
    end


This solved our update problem and we are using it in production for our app.
Using the `in_batches` API, the updates are done in smaller batches.
The database can process other queries as well without any problem.

To try this feature upgrade to the newest Rails commit or use this [gem](https://github.com/siadat/in_batches) in your existing app.
