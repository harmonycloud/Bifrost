#include <linux/module.h>
#include <linux/init.h>
#include <linux/kernel.h>
#include <linux/types.h>
#include <linux/netdevice.h>
#include <linux/inetdevice.h>

#include <linux/string.h>

 static void send_sig_func(void);

int recv_netdev_event(struct notifier_block *this, unsigned long event, void *ptr)
{
   struct net_device *dev = (struct net_device *)ptr;

    switch(event)
    {
        case NETDEV_NOTIFY_PEERS:
             if(dev && dev->name && strcmp(dev->name,"bond0") == 0){
                 printk("dev[%s] is up\n",dev->name);
                 send_sig_func();
             }

        default:
             if(dev && dev->name)
                 printk("dev[%s] ,event[%lu]\n",dev->name,event);
             break;
    }

    return NOTIFY_DONE;
}

static void send_sig_func(void){
    int ret;
    struct task_struct *task=&init_task;
    struct task_struct *p;
    struct list_head *pos;
    struct siginfo info;
    int sig_num = SIGUSR1;
    char pro_name[] = "garp";
    memset(&info,0,sizeof(struct siginfo));
    info.si_signo = sig_num;
    info.si_code = SI_QUEUE;
    info.si_int = 1234;
    list_for_each(pos,&task->tasks){
        p = list_entry(pos, struct task_struct,tasks);
        if(strcmp(p->comm, pro_name) == 0){
            printk(KERN_DEBUG "find %s: %d ----> %s, send sig: %d\n", pro_name, p->pid, p->comm,sig_num);
            ret = send_sig_info(sig_num, &info , p);
            if(ret <0){
                printk(KERN_DEBUG "error sending signal\n");
            }
            break;
        }
    }
}

struct notifier_block devhandle={
    .notifier_call = recv_netdev_event
};

static int __init  ha_init(void)
{
    register_netdevice_notifier(&devhandle);

    return 0;
}

static void __exit ha_exit(void)
{
    unregister_inetaddr_notifier(&devhandle);
    return;
}




module_init(ha_init);
module_exit(ha_exit);

MODULE_LICENSE("GPL");


