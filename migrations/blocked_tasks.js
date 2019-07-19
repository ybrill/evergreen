// constants
finishedStatuses = ["success","failed"]
startedStatuses = ["started", "dispatched"]

function taskIsBlocked(task) {
    for(dep of task.depends_on) {
        if (dep.unattainable) {
            return true
        }
    }

    return false
}

function unattainableResolved(task) {
    for (dep of task.depends_on) {
        if (dep.unattainable == null) {
            return false
        }
    }
    return true
}

function getDepTaskMap(task) {
    var depIDs = []
    for (dep of task.depends_on) {
        depIDs.push(dep._id)
    }
    depTasks = db.tasks.find({"_id": {"$in": depIDs}}, {"status": 1, "depends_on": 1}).toArray()

    var depTaskMap = {}
    for (t of depTasks) {
        depTaskMap[t._id] = t
    }

    return depTaskMap
}

function completeVersions(tasks) {
    versionIds = new Set()
    for (task of tasks) {
        if (task.version != null){
            versionIds.add(task.version)
        }
    }
    return db.tasks.find({"version": {"$in": Array.from(versionIds)}, "status":{"$in": ["undispatched", "inactive", ""]}, "depends_on.status": {"$in": ["success", "failed", "", "*"]}, "depends_on": {"$elemMatch":{"unattainable": {"$exists": false}}}}, {"depends_on":1}).toArray()
}

//
// example invocation:
//
// sleepMilliseconds = 100
// limit = 1000
// oneWeekAgo = new Date()
// oneWeekAgo.setDate(oneWeekAgo.getDate() - 7)
// migrate(sleepMilliseconds, limit, oneWeekAgo)
//

function migrate(sleepMilliseconds, limit, latestDate) {
    var loops = 0
    while(true) {
        var tasks = db.tasks.find({"status":{"$in": ["undispatched", "inactive", ""]}, "depends_on.status": {"$in": ["success", "failed", "", "*"]}, "depends_on": {"$elemMatch":{"unattainable": {"$exists": false}}}, "injest_time": {"$lte": latestDate}}, {"depends_on":1, "version":1}).limit(limit).toArray()
        tasks = completeVersions(tasks)
        if (tasks.length == 0) {
            printjson("finished")
            break
        }
        printjson("Loop number: " + loops++)

        var tasksUpdated = 0
        for (i=0; i < tasks.length; i++) {
            taskUpdated = false
            dependsOn = tasks[i].depends_on
            depTasksMap = getDepTaskMap(tasks[i])
            for (j=0; j < dependsOn.length; j++) {
                if(dependsOn[j].unattainable != null) {
                    continue
                }

                depTask = depTasksMap[dependsOn[j]._id]
                
                if (dependsOn[j]._id == tasks[i]._id) {
                    taskUpdated = true
                    dependsOn[j].unattainable = true
                } else if (!(dependsOn[j]._id in depTasksMap)) {
                    taskUpdated = true
                    dependsOn[j].unattainable = true
                } else if (finishedStatuses.includes(depTask.status)) {
                    taskUpdated = true
                    // 1st degree blocked
                    if(dependsOn[j].status != "*" && depTask.status != dependsOn[j].status) {
                        dependsOn[j].unattainable = true
                    } else {
                        dependsOn[j].unattainable = false
                    }
                } else if (startedStatuses.includes(depTask.status)) {
                    taskUpdated = true
                    dependsOn[j].unattainable = false
                } else if (dependsOn[j].status != "*" && taskIsBlocked(depTask)) {
                    taskUpdated = true
                    dependsOn[j].unattainable = true
                } else if (unattainableResolved(depTask)) {
                    taskUpdated = true
                    dependsOn[j].unattainable = false
                }
            }
            if(taskUpdated) {
                tasksUpdated++
                db.tasks.updateOne({"_id": tasks[i]._id}, {"$set": {"depends_on": dependsOn}})
            }
        }
        if(tasksUpdated == 0) {
            printjson("no tasks updated")
            break
        }
        printjson("Tasks updated:" + tasksUpdated)
        sleep(sleepMilliseconds)
    }
}
