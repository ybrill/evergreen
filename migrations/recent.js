function clearUnattainable(limit, sleepMs) {
    var totalUpdated = 0
    var ids = []

    do {
        let tasksToUpdate = db.tasks.find({"create_time": {"$gte": ISODate("2019-08-01T00:00:00.000Z")}, "injest_time": {"$gte": ISODate("2019-08-27T00:00:00.000Z"), "$lte": ISODate("2019-09-04T14:00:00.000Z")}, "depends_on.unattainable": {"$exists": true}}).limit(limit).toArray()
        ids = tasksToUpdate.map(task => task._id)
        // Update every document in the array of nested depends_on documents
        // https://docs.mongodb.com/manual/reference/operator/update/positional-all/#update-all-documents-in-an-array
        db.tasks.updateMany({"_id": {"$in": ids}}, {"$unset": {"depends_on.$[].unattainable": ""}})

        totalUpdated += ids.length
        printjson("updated " + ids.length)

        sleep(sleepMs)
    } while(ids.length > 0)

    printjson("total updated: " + totalUpdated)
}