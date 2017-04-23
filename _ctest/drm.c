#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/ioctl.h>

#define DRM_IOCTL_NR(n)		_IOC_NR(n)
#define DRM_IOC_VOID		_IOC_NONE
#define DRM_IOC_READ		_IOC_READ
#define DRM_IOC_WRITE		_IOC_WRITE
#define DRM_IOC_READWRITE	_IOC_READ|_IOC_WRITE
#define DRM_IOC(dir, group, nr, size) _IOC(dir, group, nr, size
#define DRM_IOCTL_BASE			'd'
#define DRM_IO(nr)			_IO(DRM_IOCTL_BASE,nr)
#define DRM_IOR(nr,type)		_IOR(DRM_IOCTL_BASE,nr,type)
#define DRM_IOW(nr,type)		_IOW(DRM_IOCTL_BASE,nr,type)
#define DRM_IOWR(nr,type)		_IOWR(DRM_IOCTL_BASE,nr,type)

#define DRM_IOCTL_VERSION		DRM_IOWR(0x00, struct _drmVersion)

typedef struct _drmVersion {
    int     version_major;        /**< Major version */
    int     version_minor;        /**< Minor version */
    int     version_patchlevel;   /**< Patch level */
    int     name_len; 	          /**< Length of name buffer */
    char    *name;	          /**< Name of driver */
    int     date_len;             /**< Length of date buffer */
    char    *date;                /**< User-space buffer to hold date */
    int     desc_len;	          /**< Length of desc buffer */
    char    *desc;                /**< User-space buffer to hold desc */
} drmVersion, *drmVersionPtr;

struct drm_version {
	int version_major;	  /**< Major version */
	int version_minor;	  /**< Minor version */
	int version_patchlevel;	  /**< Patch level */
	size_t name_len;	  /**< Length of name buffer */
	char *name;	  /**< Name of driver */
	size_t date_len;	  /**< Length of date buffer */
	char *date;	  /**< User-space buffer to hold date */
	size_t desc_len;	  /**< Length of desc buffer */
	char *desc;	  /**< User-space buffer to hold desc */
};

typedef struct drm_version drm_version_t;

int main() {
        int fd = open("/dev/dri/card0", O_RDWR, 0);
        if(fd < 0) {
                return 1;
        }

        drm_version_t *version = malloc(sizeof(*version));
        version->name_len    = 0;
        version->name        = NULL;
        version->date_len    = 0;
        version->date        = NULL;
        version->desc_len    = 0;
        version->desc        = NULL;

        int err = ioctl(fd, DRM_IOCTL_VERSION, version);
        if(err == 0) {
                printf("success: %d %d %d %d\n", version->version_major, version->version_minor, version->version_patchlevel, version->date_len);
        } else {
                printf("failed: %d\n", err);
                return 1;
        }
        return 0;

}
