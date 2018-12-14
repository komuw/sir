# DOCS:
# 1. https://www.google.com/url?q=https%3A%2F%2Fscikit-learn.org%2Fstable%2Fauto_examples%2Fcluster%2Fplot_dbscan.html%23sphx-glr-auto-examples-cluster-plot-dbscan-py&sa=D&sntz=1&usg=AFQjCNG5knjPH6-c706IZxq6V7DfZrFJVg
# 2. https://www.youtube.com/watch?v=vN5dAZrS58E

import numpy as np

from sklearn.cluster import DBSCAN
from sklearn import metrics
from sklearn.datasets.samples_generator import make_blobs
from sklearn.preprocessing import StandardScaler


# #############################################################################
# Generate sample data
centers = [[1, 1], [-1, -1], [1, -1]]
X, labels_true = make_blobs(n_samples=750, centers=centers, cluster_std=0.4, random_state=0)

X = StandardScaler().fit_transform(X)

# Compute DBSCAN
X = np.array([[1, 2], [2, 2], [2, 3], [8, 7], [8, 8], [25, 80]])
# X = np.array([[1], [2], [3], [7], [8], [80], [123]])
db = DBSCAN(eps=3, min_samples=2).fit(X)


# db = DBSCAN(eps=0.3, min_samples=10).fit(X)
core_samples_mask = np.zeros_like(db.labels_, dtype=bool)
core_samples_mask[db.core_sample_indices_] = True
labels = db.labels_


# Number of clusters in labels, ignoring noise if present.
n_clusters_ = len(set(labels)) - (1 if -1 in labels else 0)
n_noise_ = list(labels).count(-1)


print("Estimated number of clusters: %d" % n_clusters_)
print("Estimated number of noise points: %d" % n_noise_)
# print("Homogeneity: %0.3f" % metrics.homogeneity_score(labels_true, labels))
# print("Completeness: %0.3f" % metrics.completeness_score(labels_true, labels))
# print("V-measure: %0.3f" % metrics.v_measure_score(labels_true, labels))
# print("Adjusted Rand Index: %0.3f" % metrics.adjusted_rand_score(labels_true, labels))
# print(
#     "Adjusted Mutual Information: %0.3f" % metrics.adjusted_mutual_info_score(labels_true, labels)
# )
# print("Silhouette Coefficient: %0.3f" % metrics.silhouette_score(X, labels))


def find_cluster_members(labels, X):
    """
    for:
        X = np.array([[1], [2], [3], [7], [8], [80], [123]])
    
        labels == array([ 0,  0,  0,  1,  1, -1, -1])
    And Number of elements in this array(labels) equals number of rows, ie: len(labels) == len(X)
    If first element(of array labels) is 5 it means that row 1(of X) belongs to cluster 5.

    Thus:
      cluster_members {'noise': [array([80]), array([123])], '0': [array([1]), array([2]), array([3])], '1': [array([7]), array([8])]}
    

    Also, the following:
        z1 = bytearray("komu", encoding="utf8")
        z2 = bytearray("nomu", encoding="utf8")
        z3 = bytearray("iomu", encoding="utf8")
        z4 = bytearray("komr", encoding="utf8")
        z5 = bytearray("komt", encoding="utf8")
        z6 = bytearray("komx", encoding="utf8")
        z7 = bytearray("komg", encoding="utf8")

        X = np.array([z1, z2, z3, z4, z5, z6, z7])
    will produce this cluster:
        Estimated number of clusters: 1
        Estimated number of noise points: 1
        cluster_members {
                         'noise': [array([107, 111, 109, 103], dtype=uint8)],
                         '0': [
                               array([107, 111, 109, 117], dtype=uint8),
                               array([110, 111, 109, 117], dtype=uint8),
                               array([105, 111, 109, 117], dtype=uint8),
                               array([107, 111, 109, 114], dtype=uint8),
                               array([107, 111, 109, 116], dtype=uint8),
                               array([107, 111, 109, 120], dtype=uint8)
                            ]
                        }
    where noise': [array([107, 111, 109, 103], dtype=uint8)], is bytearray("komg", encoding="utf8")
    this is because `g` is not within eps=3 of `u` or any of the other letters used as the 4th character ie `r` `t` and `x`
    but if we replaced `g` with `p` there would be no noise.
    This is because `p` is within eps=3 of `r` which is within eps=3 of `t` which is within eps=3 of the `u` found in z1="komu"
    """

    assert len(labels) == len(X)
    cluster_members = {"noise": []}
    for val in labels:
        if val != -1:
            cluster_members.update({str(val): []})

    for k, v in enumerate(labels):
        if v == -1:
            # noise
            cluster_members["noise"].append(X[k])
        for cluster_members_key in cluster_members.keys():
            if str(v) == cluster_members_key:
                cluster_members[cluster_members_key].append(X[k])
    return cluster_members


cluster_members = find_cluster_members(labels=labels, X=X)
print("cluster_members", cluster_members)

# #############################################################################
# Plot result
# import matplotlib.pyplot as plt

# # Black removed and is used for noise instead.
# unique_labels = set(labels)
# colors = [plt.cm.Spectral(each) for each in np.linspace(0, 1, len(unique_labels))]
# for k, col in zip(unique_labels, colors):
#     if k == -1:
#         # Black used for noise.
#         col = [0, 0, 0, 1]

#     class_member_mask = labels == k

#     xy = X[class_member_mask & core_samples_mask]
#     plt.plot(
#         xy[:, 0], xy[:, 1], "o", markerfacecolor=tuple(col), markeredgecolor="k", markersize=14
#     )

#     xy = X[class_member_mask & ~core_samples_mask]
#     plt.plot(xy[:, 0], xy[:, 1], "o", markerfacecolor=tuple(col), markeredgecolor="k", markersize=6)

# plt.title("Estimated number of clusters: %d" % n_clusters_)
# plt.show()
